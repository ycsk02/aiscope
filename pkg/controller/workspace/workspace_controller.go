/*
Copyright 2021.

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

package workspace

import (
	tenantv1alpha2 "aiscope/pkg/apis/tenant/v1alpha2"
	"aiscope/pkg/utils/k8sutil"
	"aiscope/pkg/utils/sliceutil"
	"context"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	controllerutils "aiscope/pkg/controller/utils/controller"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	controllerName = "workspace-controller"
)

// Reconciler reconciles a Workspace object
type Reconciler struct {
	client.Client
	Logger                  logr.Logger
	Recorder                record.EventRecorder
	MaxConcurrentReconciles int
}

//+kubebuilder:rbac:groups=tenant.aiscope.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tenant.aiscope.io,resources=workspaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tenant.aiscope.io,resources=workspaces/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Workspace object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.WithValues("workspace", req.NamespacedName)
	rootCtx := context.Background()
	workspace := &tenantv1alpha2.Workspace{}
	if err := r.Get(rootCtx, req.NamespacedName, workspace); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	finalizer := "finalizers.aiscope.io/tenant"

	if workspace.ObjectMeta.DeletionTimestamp.IsZero() {
		if !sliceutil.HasString(workspace.ObjectMeta.Finalizers, finalizer) {
			workspace.ObjectMeta.Finalizers = append(workspace.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(rootCtx, workspace); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if sliceutil.HasString(workspace.ObjectMeta.Finalizers, finalizer) {
			workspace.ObjectMeta.Finalizers = sliceutil.RemoveString(workspace.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})
			logger.V(4).Info("update workspace")
			if err := r.Update(rootCtx,workspace); err != nil {
				logger.Error(err, "update workspace failed")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	var namespaces corev1.NamespaceList
	if err := r.List(rootCtx, &namespaces, client.MatchingLabels{tenantv1alpha2.WorkspaceLabel: req.Name}); err != nil {
		logger.Error(err, "list namespaces failed")
		return ctrl.Result{}, err
	} else {
		for _, namespace := range namespaces.Items {
			if err := r.bindWorkspace(rootCtx, logger, &namespace, workspace); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	r.Recorder.Event(workspace, corev1.EventTypeNormal, controllerutils.SuccessSynced, controllerutils.MessageResourceSynced)
	return ctrl.Result{}, nil
}

func (r *Reconciler) bindWorkspace(ctx context.Context, logger logr.Logger, namespace *corev1.Namespace, workspace *tenantv1alpha2.Workspace) error {
	// owner reference not match workspace label
	if !metav1.IsControlledBy(namespace, workspace) {
		namespace := namespace.DeepCopy()
		namespace.OwnerReferences = k8sutil.RemoveWorkspaceOwnerReference(namespace.OwnerReferences)
		if err := controllerutil.SetControllerReference(workspace, namespace, scheme.Scheme); err != nil {
			logger.Error(err, "set controller reference failed")
			return err
		}
		logger.V(4).Info("update namespace owner reference", "workspace", workspace.Name)
		if err := r.Update(ctx, namespace); err != nil {
			logger.Error(err, "update namespace failed")
			return err
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}

	r.Logger = ctrl.Log.WithName("controllers").WithName(controllerName)

	if r.Recorder == nil {
		r.Recorder = mgr.GetEventRecorderFor(controllerName)
	}
	if r.MaxConcurrentReconciles <= 0 {
		r.MaxConcurrentReconciles = 1
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.MaxConcurrentReconciles,
		}).
		For(&tenantv1alpha2.Workspace{}).
		Complete(r)
}
