/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package group

import (
	tenantv1alpha2 "aiscope/pkg/apis/tenant/v1alpha2"
	aiscope "aiscope/pkg/client/clientset/versioned"
	"aiscope/pkg/constants"
	"aiscope/pkg/controller/utils/controller"
	"aiscope/pkg/utils/k8sutil"
	"aiscope/pkg/utils/sliceutil"
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"

	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
	iamv1alpha2informers "aiscope/pkg/client/informers/externalversions/iam/v1alpha2"
	iamv1alpha2listers "aiscope/pkg/client/listers/iam/v1alpha2"

	"k8s.io/client-go/kubernetes/scheme"
)

const (
	successSynced         = "Synced"
	messageResourceSynced = "Group synced successfully"
	controllerName        = "group-controller"
	finalizer             = "finalizers.aiscope.io/groups"
)

type Controller struct {
	controller.BaseController
	scheme               *runtime.Scheme
	k8sClient            kubernetes.Interface
	aiClient             aiscope.Interface
	groupInformer        iamv1alpha2informers.GroupInformer
	groupLister          iamv1alpha2listers.GroupLister
	recorder             record.EventRecorder
}

// NewController creates Group Controller instance
func NewController(k8sClient kubernetes.Interface, aiClient aiscope.Interface, groupInformer iamv1alpha2informers.GroupInformer) *Controller {

	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	ctl := &Controller{
		BaseController: controller.BaseController{
			Workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Group"),
			Synced:    []cache.InformerSynced{groupInformer.Informer().HasSynced},
			Name:      controllerName,
		},
		recorder:            eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName}),
		k8sClient:           k8sClient,
		aiClient:            aiClient,
		groupInformer:       groupInformer,
		groupLister:         groupInformer.Lister(),
	}

	ctl.Handler = ctl.reconcile

	klog.Info("Setting up event handlers")
	groupInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.Enqueue,
		UpdateFunc: func(old, new interface{}) {
			ctl.Enqueue(new)
		},
		DeleteFunc: ctl.Enqueue,
	})
	return ctl
}

func (c *Controller) Start(ctx context.Context) error {
	return c.Run(1, ctx.Done())
}

// reconcile handles Group informer events, clear up related reource when group is being deleted.
func (c *Controller) reconcile(key string) error {

	group, err := c.groupLister.Get(key)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("group '%s' in work queue no longer exists", key))
			return nil
		}
		klog.Error(err)
		return err
	}

	if group.ObjectMeta.DeletionTimestamp.IsZero() {
		var g *iamv1alpha2.Group
		if !sliceutil.HasString(group.Finalizers, finalizer) {
			g = group.DeepCopy()
			g.ObjectMeta.Finalizers = append(g.ObjectMeta.Finalizers, finalizer)
		}

		// Set OwnerReferences when the group has a parent or Workspace.
		if group.Labels != nil {
			if parent, ok := group.Labels[iamv1alpha2.GroupParent]; ok {
				// If the Group is owned by a Parent
				if !k8sutil.IsControlledBy(group.OwnerReferences, "Group", parent) {
					if g == nil {
						g = group.DeepCopy()
					}
					groupParent, err := c.groupLister.Get(parent)
					if err != nil {
						if errors.IsNotFound(err) {
							utilruntime.HandleError(fmt.Errorf("Parent group '%s' no longer exists", key))
							delete(g.Labels, iamv1alpha2.GroupParent)
						} else {
							klog.Error(err)
							return err
						}
					} else {
						if err := controllerutil.SetControllerReference(groupParent, g, scheme.Scheme); err != nil {
							klog.Error(err)
							return err
						}
					}
				}
			} else if ws, ok := group.Labels[constants.WorkspaceLabelKey]; ok {
				// If the Group is owned by a Workspace
				if !k8sutil.IsControlledBy(group.OwnerReferences, tenantv1alpha2.ResourceKindWorkspace, ws) {
					workspace, err := c.aiClient.TenantV1alpha2().Workspaces().Get(context.Background(), ws, metav1.GetOptions{})
					if err != nil {
						if errors.IsNotFound(err) {
							utilruntime.HandleError(fmt.Errorf("Workspace '%s' no longer exists", ws))
						} else {
							klog.Error(err)
							return err
						}
					} else {
						if g == nil {
							g = group.DeepCopy()
						}
						g.OwnerReferences = k8sutil.RemoveWorkspaceOwnerReference(g.OwnerReferences)
						if err := controllerutil.SetControllerReference(workspace, g, scheme.Scheme); err != nil {
							return err
						}
					}
				}
			}
		}
		if g != nil {
			if _, err = c.aiClient.IamV1alpha2().Groups().Update(context.Background(), g, metav1.UpdateOptions{}); err != nil {
				return err
			}
			// Skip reconcile when group is updated.
			return nil
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(group.ObjectMeta.Finalizers, finalizer) {
			if err = c.deleteGroupBindings(group); err != nil {
				klog.Error(err)
				return err
			}

			if err = c.deleteRoleBindings(group); err != nil {
				klog.Error(err)
				return err
			}

			group.Finalizers = sliceutil.RemoveString(group.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})

			if group, err = c.aiClient.IamV1alpha2().Groups().Update(context.Background(), group, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
		return nil
	}

	c.recorder.Event(group, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return nil
}

func (c *Controller) deleteGroupBindings(group *iamv1alpha2.Group) error {
	if len(group.Name) > validation.LabelValueMaxLength {
		// ignore invalid label value error
		return nil
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(labels.Set{iamv1alpha2.GroupReferenceLabel: group.Name}).String(),
	}
	if err := c.aiClient.IamV1alpha2().GroupBindings().
		DeleteCollection(context.Background(), *metav1.NewDeleteOptions(0), listOptions); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

// remove all RoleBindings.
func (c *Controller) deleteRoleBindings(group *iamv1alpha2.Group) error {
	if len(group.Name) > validation.LabelValueMaxLength {
		// ignore invalid label value error
		return nil
	}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(labels.Set{iamv1alpha2.GroupReferenceLabel: group.Name}).String(),
	}
	deleteOptions := *metav1.NewDeleteOptions(0)

	if err := c.aiClient.IamV1alpha2().WorkspaceRoleBindings().
		DeleteCollection(context.Background(), deleteOptions, listOptions); err != nil {
		klog.Error(err)
		return err
	}

	if err := c.k8sClient.RbacV1().ClusterRoleBindings().
		DeleteCollection(context.Background(), deleteOptions, listOptions); err != nil {
		klog.Error(err)
		return err
	}

	if result, err := c.k8sClient.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{}); err != nil {
		klog.Error(err)
		return err
	} else {
		for _, namespace := range result.Items {
			if err = c.k8sClient.RbacV1().RoleBindings(namespace.Name).DeleteCollection(context.Background(), deleteOptions, listOptions); err != nil {
				klog.Error(err)
				return err
			}
		}
	}

	return nil
}

