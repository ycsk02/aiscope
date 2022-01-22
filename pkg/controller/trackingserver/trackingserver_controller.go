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
	"aiscope/pkg/utils/sliceutil"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"net/url"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	experimentv1alpha2 "aiscope/pkg/apis/experiment/v1alpha2"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	traefikclient "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	networkv1 "k8s.io/api/networking/v1"
)

const (
	successSynced = "Synced"
	failedSynced  = "FailedSync"
	// is synced successfully
	messageResourceSynced = "Trackingserver synced successfully"
	controllerName        = "trackingserver-controller"
	finalizer       = "finalizers.aiscope.io/trackingserver"
	interval        = time.Second
	timeout         = 15 * time.Second
	syncFailMessage = "Failed to sync: %s"
)

// TrackingServerReconciler reconciles a TrackingServer object
type TrackingServerReconciler struct {
	client.Client
	TraefikClient           traefikclient.Interface
	Logger                  logr.Logger
	Recorder                record.EventRecorder
	IngressController       string
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
	logger := r.Logger.WithValues("trackingServer", req.NamespacedName)
	rootCtx := context.Background()

	trackingServer := &experimentv1alpha2.TrackingServer{}
	if err := r.Get(rootCtx, req.NamespacedName, trackingServer); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if trackingServer.ObjectMeta.DeletionTimestamp.IsZero() {
		if !sliceutil.HasString(trackingServer.ObjectMeta.Finalizers, finalizer) {
			trackingServer.ObjectMeta.Finalizers = append(trackingServer.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(rootCtx, trackingServer); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if sliceutil.HasString(trackingServer.ObjectMeta.Finalizers, finalizer) {
			trackingServer.ObjectMeta.Finalizers = sliceutil.RemoveString(trackingServer.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})
			logger.V(4).Info("update trackingServer")
			if err := r.Update(rootCtx,trackingServer); err != nil {
				logger.Error(err, "update trackingServer failed")
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if err := r.reconcileSecret(rootCtx, logger, trackingServer); err != nil {
		klog.Error(err)
		r.Recorder.Event(trackingServer, corev1.EventTypeWarning, failedSynced, fmt.Sprintf(syncFailMessage, err))
		return reconcile.Result{}, err
	}

	if err := r.reconcileDeployment(rootCtx, logger, trackingServer); err != nil {
		klog.Error(err)
		r.Recorder.Event(trackingServer, corev1.EventTypeWarning, failedSynced, fmt.Sprintf(syncFailMessage, err))
		return reconcile.Result{}, err
	}

	if err := r.reconcileService(rootCtx, logger, trackingServer); err != nil {
		klog.Error(err)
		r.Recorder.Event(trackingServer, corev1.EventTypeWarning, failedSynced, fmt.Sprintf(syncFailMessage, err))
		return reconcile.Result{}, err
	}

	if err := r.reconcileIngress(rootCtx, logger, trackingServer); err != nil {
		klog.Error(err)
		r.Recorder.Event(trackingServer, corev1.EventTypeWarning, failedSynced, fmt.Sprintf(syncFailMessage, err))
		return reconcile.Result{}, err
	}

	r.Recorder.Event(trackingServer, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return ctrl.Result{}, nil
}

func (r *TrackingServerReconciler) reconcileDeployment(ctx context.Context, logger logr.Logger, instance *experimentv1alpha2.TrackingServer) error {
	expectDployment := newDeploymentForTrackingServer(instance)
	if err := controllerutil.SetControllerReference(instance, expectDployment, scheme.Scheme); err != nil {
		logger.Error(err, "set controller reference failed")
		return err
	}

	currentDeployment := &appsv1.Deployment{}
	 if err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, currentDeployment); err != nil {
		 if errors.IsNotFound(err) {
			 logger.V(4).Info("create trackingserver deployment", "trackingserver", instance.Name)
			 if err := r.Create(ctx, expectDployment); err != nil {
				 logger.Error(err, "create trackingserver deployment  failed")
				 return err
			 }
			 return nil
		 }

		 logger.Error(err, "get trackingserver deployment failed")
		 return err
	 } else {
		 if !reflect.DeepEqual(expectDployment.Spec, currentDeployment.Spec) {
			 currentDeployment.Spec = expectDployment.Spec
			 logger.V(4).Info("update trackingserver deployment", "trackingserver", instance.Name)
			 if err := r.Update(ctx, currentDeployment); err != nil {
				 logger.Error(err, "update trackingserver deployment failed")
				 return err
			 }
		 }
	 }

	 return nil
}

func newDeploymentForTrackingServer(instance *experimentv1alpha2.TrackingServer) *appsv1.Deployment {
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

func (r *TrackingServerReconciler) reconcileService(ctx context.Context, logger logr.Logger, instance *experimentv1alpha2.TrackingServer) error {
	expectService := newServiceForTrackingServer(instance)
	if err := controllerutil.SetControllerReference(instance, expectService, scheme.Scheme); err != nil {
		logger.Error(err, "set controller reference failed")
		return err
	}

	currentService := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, currentService); err != nil {
		if errors.IsNotFound(err) {
			logger.V(4).Info("create trackingserver service", "trackingserver", instance.Name)
			if err := r.Create(ctx, expectService); err != nil {
				logger.Error(err, "create trackingserver service failed")
				return err
			}
			return nil
		}

		logger.Error(err, "get trackingserver service failed")
		return err
	} else {
		if !reflect.DeepEqual(expectService.Spec, currentService.Spec) {
			currentService.Spec = expectService.Spec
			logger.V(4).Info("update trackingserver service", "trackingserver", instance.Name)
			if err := r.Update(ctx, currentService); err != nil {
				logger.Error(err, "update trackingserver service failed")
				return err
			}
		}
	}

	return nil
}

func newServiceForTrackingServer(instance *experimentv1alpha2.TrackingServer) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:       instance.Name,
			Namespace:  instance.Namespace,
			Labels:     labelsForTrackingServer(instance.Name),
		},
		Spec: corev1.ServiceSpec{
			Type:       corev1.ServiceTypeClusterIP,
			Ports:      []corev1.ServicePort{
				{
					Name:           "server",
					Port:           5000,
					TargetPort:     intstr.FromInt(5000),
				},
			},
			Selector:   labelsForTrackingServer(instance.Name),
		},
	}

	return service
}

func labelsForTrackingServer(name string) map[string]string {
	return map[string]string{"app": "trackingserver", "ts_name": name}
}

func (r *TrackingServerReconciler) reconcileSecret(ctx context.Context, logger logr.Logger, instance *experimentv1alpha2.TrackingServer) error {
	if instance.Spec.Cert == "" || instance.Spec.Key == "" {
		return nil
	}

	expectSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:       instance.Name,
			Namespace:  instance.Namespace,
		},
		Data: map[string][]byte{
			"tls.crt":  []byte(instance.Spec.Cert),
			"tls.key":  []byte(instance.Spec.Key),
		},
		Type: corev1.SecretTypeTLS,
	}

	if err := controllerutil.SetControllerReference(instance, expectSecret, scheme.Scheme); err != nil {
		logger.Error(err, "set controller reference failed")
		return err
	}

	currentSecret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, currentSecret); err != nil {
		if errors.IsNotFound(err) {
			logger.V(4).Info("create trackingserver secret", "trackingserver", instance.Name)
			if err := r.Create(ctx, expectSecret); err != nil {
				logger.Error(err, "create trackingserver secret failed")
				return err
			}
			return nil
		}

		logger.Error(err, "get trackingserver secret failed")
		return err
	} else {
		if !reflect.DeepEqual(expectSecret.Data, currentSecret.Data) {
			currentSecret.Data = expectSecret.Data
			logger.V(4).Info("update trackingserver secret", "trackingserver", instance.Name)
			if err := r.Update(ctx, currentSecret); err != nil {
				logger.Error(err, "update trackingserver secret failed")
				return err
			}
		}
	}

	return nil
}

func (r *TrackingServerReconciler) reconcileIngress(ctx context.Context, logger logr.Logger, instance *experimentv1alpha2.TrackingServer) error {
	secret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, secret); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		secret = nil
	}

	if r.IngressController == "traefik" {
		if err := r.reconcileTraefikRoute(ctx, logger, instance, secret); err != nil {
			return err
		}
		return nil
	}

	if r.IngressController == "nginx" {
		if err := r.reconcileNginxIngress(ctx, logger, instance, secret); err != nil {
			return err
		}
		return nil
	}

	return nil
}

func (r *TrackingServerReconciler) reconcileTraefikRoute(ctx context.Context, logger logr.Logger, instance *experimentv1alpha2.TrackingServer, secret *corev1.Secret) error {
	parsedUrl, err := url.Parse(instance.Spec.URL)
	if err != nil {
		logger.Error(err, "parse url failed")
		return err
	}

	expectIngressRoute := ingressRouteForTrackingServer(instance, secret, parsedUrl)
	if err := controllerutil.SetControllerReference(instance, expectIngressRoute, scheme.Scheme); err != nil {
		logger.Error(err, "set controller reference failed")
		return err
	}
	currentIngressRoute, err := r.TraefikClient.TraefikV1alpha1().IngressRoutes(instance.Namespace).Get(ctx, instance.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = r.TraefikClient.TraefikV1alpha1().IngressRoutes(instance.Namespace).Create(ctx, expectIngressRoute, metav1.CreateOptions{})
			if err != nil {
				klog.Error(err)
				return err
			}
		} else {
			return err
		}
	} else {
		if !reflect.DeepEqual(expectIngressRoute.Spec, currentIngressRoute.Spec) {
			currentIngressRoute.Spec = expectIngressRoute.Spec
			_, err = r.TraefikClient.TraefikV1alpha1().IngressRoutes(instance.Namespace).Update(ctx, currentIngressRoute, metav1.UpdateOptions{})
			if err != nil {
				klog.Error(err)
				return err
			}
		}
	}

	expectMiddleware := middlewareForTrackingServer(instance, parsedUrl)
	if err := controllerutil.SetControllerReference(instance, expectMiddleware, scheme.Scheme); err != nil {
		logger.Error(err, "set controller reference failed")
		return err
	}
	currentMiddleware, err := r.TraefikClient.TraefikV1alpha1().Middlewares(instance.Namespace).Get(ctx, instance.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = r.TraefikClient.TraefikV1alpha1().Middlewares(instance.Namespace).Create(ctx, expectMiddleware, metav1.CreateOptions{})
			if err != nil {
				klog.Error(err)
				return err
			}
		} else {
			klog.Error(err)
			return err
		}
	} else {
		if !reflect.DeepEqual(expectMiddleware.Spec, currentMiddleware.Spec) {
			currentMiddleware.Spec = expectMiddleware.Spec
			_, err = r.TraefikClient.TraefikV1alpha1().Middlewares(instance.Namespace).Update(ctx, currentMiddleware, metav1.UpdateOptions{})
			if err != nil {
				klog.Error(err)
				return err
			}
		}
	}

	return nil
}

func (r *TrackingServerReconciler) reconcileNginxIngress(ctx context.Context, logger logr.Logger, instance *experimentv1alpha2.TrackingServer, secret *corev1.Secret) error {

	parsedUrl, err := url.Parse(instance.Spec.URL)
	if err != nil {
		logger.Error(err, "parse url failed")
		return err
	}

	currentSecret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, currentSecret); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}

	expectIngress := newIngressForTrackingServer(instance, secret, parsedUrl)

	if err := controllerutil.SetControllerReference(instance, expectIngress, scheme.Scheme); err != nil {
		logger.Error(err, "set controller reference failed")
		return err
	}

	currentIngress := &networkv1.Ingress{}
	if err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, currentIngress); err != nil {
		if errors.IsNotFound(err) {
			logger.V(4).Info("create trackingserver", "trackingserver", instance.Name)
			if err := r.Create(ctx, expectIngress); err != nil {
				logger.Error(err, "create trackingserver failed")
				return err
			}
			return nil
		}

		logger.Error(err, "get trackingserver failed")
		return err
	} else {
		if !reflect.DeepEqual(expectIngress.Spec, currentIngress.Spec) || !reflect.DeepEqual(expectIngress.Annotations, currentIngress.Annotations) {
			currentIngress.Spec = expectIngress.Spec
			logger.V(4).Info("update trackingserver", "trackingserver", instance.Name)
			if err := r.Update(ctx, currentIngress); err != nil {
				logger.Error(err, "update trackingserver failed")
				return err
			}
		}
	}

	return nil
}

func newIngressForTrackingServer(instance *experimentv1alpha2.TrackingServer, secret *corev1.Secret, parsedUrl *url.URL) *networkv1.Ingress {
	pathType := networkv1.PathTypeImplementationSpecific
	ingress := &networkv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:       instance.Name,
			Namespace:  instance.Namespace,
		},
		Spec: networkv1.IngressSpec{
			Rules: []networkv1.IngressRule{
				{},
			},
		},
	}

	if ingress.Annotations == nil {
		ingress.Annotations = make(map[string]string)
	}

	if parsedUrl.Path != "" {
		ingress.Annotations["nginx.ingress.kubernetes.io/rewrite-target"] = "/$2"
		ingress.Spec.Rules = []networkv1.IngressRule{
			{
				Host:               parsedUrl.Host,
				IngressRuleValue:   networkv1.IngressRuleValue{
					HTTP: &networkv1.HTTPIngressRuleValue{
						Paths: []networkv1.HTTPIngressPath{
							{
								Path: fmt.Sprintf("%s(/|$)(.*)", parsedUrl.Path),
								PathType: &pathType,
								Backend: networkv1.IngressBackend{
									Service: &networkv1.IngressServiceBackend{
										Name: instance.Name,
										Port: networkv1.ServiceBackendPort{
											Name: "server",
										},
									},
								},
							},
							{
								Path:   "/static-files",
								PathType: &pathType,
								Backend: networkv1.IngressBackend{
									Service: &networkv1.IngressServiceBackend{
										Name: instance.Name,
										Port: networkv1.ServiceBackendPort{
											Name: "server",
										},
									},
								},
							},
							{
								Path:   "/ajax-api",
								PathType: &pathType,
								Backend: networkv1.IngressBackend{
									Service: &networkv1.IngressServiceBackend{
										Name: instance.Name,
										Port: networkv1.ServiceBackendPort{
											Name: "server",
										},
									},
								},
							},
						},
					},
				},
			},
		}
	} else {
		ingress.Spec.Rules = []networkv1.IngressRule{
			{
				Host:               parsedUrl.Host,
				IngressRuleValue:   networkv1.IngressRuleValue{
					HTTP: &networkv1.HTTPIngressRuleValue{
						Paths: []networkv1.HTTPIngressPath{
							{
								Path:   parsedUrl.Path,
								PathType: &pathType,
								Backend: networkv1.IngressBackend{
									Service: &networkv1.IngressServiceBackend{
										Name: instance.Name,
										Port: networkv1.ServiceBackendPort{
											Name: "server",
										},
									},
								},
							},
						},
					},
				},
			},
		}

	}

	if secret != nil {
		fmt.Printf("the secret is not null %v\n", secret)
		ingress.Spec.TLS = []networkv1.IngressTLS{
			{
				Hosts: []string{
					parsedUrl.Host,
				},
				SecretName: instance.Name,
			},
		}
	}

	return ingress
}

func ingressRouteForTrackingServer(instance *experimentv1alpha2.TrackingServer, secret *corev1.Secret, parsedUrl *url.URL) *traefikv1alpha1.IngressRoute {
	ingressRoute := &traefikv1alpha1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:       instance.Name,
			Namespace:  instance.Namespace,
		},
		Spec: traefikv1alpha1.IngressRouteSpec{
			EntryPoints: []string{"web"},
		},
	}

	if secret != nil {
		ingressRoute.Spec.EntryPoints = []string{"websecure","web"}
		ingressRoute.Spec.TLS = &traefikv1alpha1.TLS{
			SecretName: instance.Name,
		}
	}

	if parsedUrl.Path == "" {
		ingressRoute.Spec.Routes = append(ingressRoute.Spec.Routes, traefikv1alpha1.Route{
			Match:  fmt.Sprintf("Host(`%s`)", parsedUrl.Host),
			Kind:   "Rule",
			Services: []traefikv1alpha1.Service{
				{
					LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{
						Name:   instance.Name,
						Port:   intstr.FromString("server"),
					},
				},
			},
		})
	} else {
		ingressRoute.Spec.Routes = append(ingressRoute.Spec.Routes, []traefikv1alpha1.Route{
			{
				Match:  fmt.Sprintf("Host(`%s`) && PathPrefix(`%s`)", parsedUrl.Host, parsedUrl.Path),
				Kind:   "Rule",
				Services: []traefikv1alpha1.Service{
					{
						LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{
							Name:   instance.Name,
							Port:   intstr.FromString("server"),
						},
					},
				},
				Middlewares: []traefikv1alpha1.MiddlewareRef{
					{
						Name: instance.Name,
					},
				},
			},
			{
				Match:  fmt.Sprintf("Host(`%s`) && (PathPrefix(`/static-files`) || PathPrefix(`/ajax-api`))", parsedUrl.Host),
				Kind:   "Rule",
				Services: []traefikv1alpha1.Service{
					{
						LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{
							Name:   instance.Name,
							Port:   intstr.FromString("server"),
						},
					},
				},
			},
		}...)
	}

	return ingressRoute
}

func middlewareForTrackingServer(instance *experimentv1alpha2.TrackingServer, parsedUrl *url.URL) *traefikv1alpha1.Middleware {
	middleware := &traefikv1alpha1.Middleware{
		ObjectMeta: metav1.ObjectMeta{
			Name:           instance.Name,
			Namespace:      instance.Namespace,
		},
		Spec: traefikv1alpha1.MiddlewareSpec{
			StripPrefix: &dynamic.StripPrefix{
				Prefixes: []string{parsedUrl.Path},
			},
		},
	}
	return middleware
}

// SetupWithManager sets up the controller with the Manager.
func (r *TrackingServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
		For(&experimentv1alpha2.TrackingServer{}).
		Complete(r)
}
