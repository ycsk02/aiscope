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

package workspacerole

import (
	tenantv1alpha2 "aiscope/pkg/apis/tenant/v1alpha2"
	"aiscope/pkg/constants"
	controllerutils "aiscope/pkg/controller/utils/controller"
	"aiscope/pkg/utils/k8sutil"
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	controllerName = "workspacerole-controller"
)

// Reconciler reconciles a WorkspaceRole object
type Reconciler struct {
	client.Client
	MultiClusterEnabled     bool
	Logger                  logr.Logger
	Scheme                  *runtime.Scheme
	Recorder                record.EventRecorder
	MaxConcurrentReconciles int
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	r.Logger = ctrl.Log.WithName("controllers").WithName(controllerName)

	if r.Scheme == nil {
		r.Scheme = mgr.GetScheme()
	}
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
		For(&iamv1alpha2.WorkspaceRole{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=iam.aiscope,resources=workspaceroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.aiscope,resources=workspaces,verbs=get;list;watch;
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.WithValues("workspacerole", req.NamespacedName)
	rootCtx := context.Background()
	workspaceRole := &iamv1alpha2.WorkspaceRole{}
	err := r.Get(rootCtx, req.NamespacedName, workspaceRole)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.bindWorkspace(rootCtx, logger, workspaceRole); err != nil {
		return ctrl.Result{}, err
	}

	r.Recorder.Event(workspaceRole, corev1.EventTypeNormal, controllerutils.SuccessSynced, controllerutils.MessageResourceSynced)
	return ctrl.Result{}, nil
}

func (r *Reconciler) bindWorkspace(ctx context.Context, logger logr.Logger, workspaceRole *iamv1alpha2.WorkspaceRole) error {
	workspaceName := workspaceRole.Labels[constants.WorkspaceLabelKey]
	if workspaceName == "" {
		return nil
	}
	var workspace tenantv1alpha2.Workspace
	if err := r.Get(ctx, types.NamespacedName{Name: workspaceName}, &workspace); err != nil {
		return client.IgnoreNotFound(err)
	}
	if !metav1.IsControlledBy(workspaceRole, &workspace) {
		workspaceRole.OwnerReferences = k8sutil.RemoveWorkspaceOwnerReference(workspaceRole.OwnerReferences)
		if err := controllerutil.SetControllerReference(&workspace, workspaceRole, r.Scheme); err != nil {
			logger.Error(err, "set controller reference failed")
			return err
		}
		if err := r.Update(ctx, workspaceRole); err != nil {
			logger.Error(err, "update workspace role failed")
			return err
		}
	}
	return nil
}
