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

package iam

import (
	aiscope "aiscope/pkg/client/clientset/versioned"
	"aiscope/pkg/controller/utils/controller"
	"context"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
	iamv1alpha2informers "aiscope/pkg/client/informers/externalversions/iam/v1alpha2"
	iamv1alpha2listers "aiscope/pkg/client/listers/iam/v1alpha2"
)

const (
	successSynced         = "Synced"
	messageResourceSynced = "Group synced successfully"
	controllerName        = "group-controller"
	finalizer             = "finalizers.kubesphere.io/groups"
)

type Controller struct {
	controller.BaseController
	scheme               *runtime.Scheme
	k8sClient            kubernetes.Interface
	aiClient             aiscope.Interface
	groupInformer        iamv1alpha2informers.GroupInformer
	groupLister          iamv1alpha2listers.GroupLister
	recorder             record.EventRecorder
	multiClusterEnabled  bool
}

// GroupReconciler reconciles a Group object
type GroupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=iam.aiscope,resources=groups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=iam.aiscope,resources=groups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=iam.aiscope,resources=groups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Group object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *GroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&iamv1alpha2.Group{}).
		Complete(r)
}
