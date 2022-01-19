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

package trackingserver

import (
	controllerutils "aiscope/pkg/controller/utils/controller"
	"aiscope/pkg/utils/sliceutil"
	"context"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	experimentv1alpha2 "aiscope/pkg/apis/experiment/v1alpha2"
)

// TrackingServerReconciler reconciles a TrackingServer object
type TrackingServerReconciler struct {
	client.Client
	Logger                  logr.Logger
	Recorder                record.EventRecorder
	MaxConcurrentReconciles int
}

//+kubebuilder:rbac:groups=experiment.aiscope,resources=trackingservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=experiment.aiscope,resources=trackingservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=experiment.aiscope,resources=trackingservers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TrackingServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *TrackingServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.WithValues("workspace", req.NamespacedName)
	rootCtx := context.Background()

	trackingServer := &experimentv1alpha2.TrackingServer{}
	if err := r.Get(rootCtx, req.NamespacedName, trackingServer); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	finalizer := "finalizers.aiscope.io/trackingserver"

	if trackingServer.ObjectMeta.DeletionTimestamp.IsZero() {
		if !sliceutil.HasString(trackingServer.ObjectMeta.Finalizers, finalizer) {
			trackingServer.ObjectMeta.Finalizers = append(trackingServer.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(rootCtx, trackingServer); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if err := r.deleteTrackingServerResource(rootCtx, logger, trackingServer); err != nil {
		}

		if sliceutil.HasString(trackingServer.ObjectMeta.Finalizers, finalizer) {
			trackingServer.ObjectMeta.Finalizers = sliceutil.RemoveString(trackingServer.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})
			logger.V(4).Info("update workspace")
			if err := r.Update(rootCtx,trackingServer); err != nil {
				logger.Error(err, "update workspace failed")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if err := r.reconcileDeployment(rootCtx, logger, trackingServer); err != nil {
		return reconcile.Result{}, err
	}

	r.Recorder.Event(trackingServer, corev1.EventTypeNormal, controllerutils.SuccessSynced, controllerutils.MessageResourceSynced)
	return ctrl.Result{}, nil
}

func (r *TrackingServerReconciler) deleteTrackingServerResource(ctx context.Context, logger logr.Logger, instance *experimentv1alpha2.TrackingServer) error {
	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, deployment); err != nil {
		if errors.IsNotFound(err) {
			logger.V(4).Info("related deployment not found", "trackingserver", instance.Name)
		} else {
			logger.Error(err, "failed to get related deployment")
		}
	} else {
		if err := r.Delete(ctx, deployment); err != nil {
			return err
		}
		logger.V(4).Info("deployment has been deleted successfully ")
	}

	service := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, service); err != nil {
		if errors.IsNotFound(err) {
			logger.V(4).Info("related service not found", "trackingserver", instance.Name)
		} else {
			logger.Error(err, "failed to get related service")
		}
	} else {
		if err := r.Delete(ctx, service); err != nil {
			return err
		}
		logger.V(4).Info("service has been deleted successfully ")
	}

	return nil
}

func (r *TrackingServerReconciler) reconcileDeployment(ctx context.Context, logger logr.Logger, instance *experimentv1alpha2.TrackingServer) error {
	expectDployment := r.newDeploymentForTrackingServer(instance)
	if err := controllerutil.SetControllerReference(instance, expectDployment, scheme.Scheme); err != nil {
		logger.Error(err, "set controller reference failed")
		return err
	}

	currentDeployment := &appsv1.Deployment{}
	 if err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, currentDeployment); err != nil {
		 if errors.IsNotFound(err) {
			 logger.V(4).Info("create trackingserver", "trackingserver", instance.Name)
			 if err := r.Create(ctx, currentDeployment); err != nil {
				 logger.Error(err, "create trackingserver failed")
				 return err
			 }
			 return nil
		 }

		 logger.Error(err, "get trackingserver failed")
		 return err
	 }

	 if !reflect.DeepEqual(expectDployment.Spec, currentDeployment.Spec) {
		 currentDeployment.Spec = expectDployment.Spec
		 logger.V(4).Info("update trackingserver", "trackingserver", instance.Name)
		 if err := r.Update(ctx, currentDeployment); err != nil {
			 logger.Error(err, "update trackingserver failed")
			 return err
		 }
	 }

	 return nil
}

func (r *TrackingServerReconciler) newDeploymentForTrackingServer(instance *experimentv1alpha2.TrackingServer) *appsv1.Deployment {
	replicas := instance.Spec.Size
	labels := labelsForTrackingServer(instance.Name)

	deployment := &appsv1.Deployment{
		ObjectMeta:     metav1.ObjectMeta{
			Name:           instance.Name,
			Namespace:      instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:       &replicas,
			Selector:       &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image:          instance.Spec.Image,
							Name:           "mlflow",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports:          []corev1.ContainerPort{{
								ContainerPort: 5000,
								Name:        "server",
							}},
							Env: []corev1.EnvVar{
								{
									Name:   "MLFLOW_TRACKING_URI",
									Value:  instance.Spec.URL,
								},
								{
									Name:   "MLFLOW_S3_ENDPOINT_URL",
									Value:  instance.Spec.S3_ENDPOINT_URL,
								},
								{
									Name:   "AWS_ACCESS_KEY_ID",
									Value:  instance.Spec.AWS_ACCESS_KEY,
								},
								{
									Name:   "AWS_SECRET_ACCESS_KEY",
									Value:  instance.Spec.AWS_SECRET_KEY,
								},
								{
									Name:   "ARTIFACT_ROOT",
									Value:  instance.Spec.ARTIFACT_ROOT,
								},
								{
									Name:   "BACKEND_URI",
									Value:  instance.Spec.BACKEND_URI,
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment
}

func labelsForTrackingServer(name string) map[string]string {
	return map[string]string{"app": "trackingserver", "ts_name": name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *TrackingServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&experimentv1alpha2.TrackingServer{}).
		Complete(r)
}
