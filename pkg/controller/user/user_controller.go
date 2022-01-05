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

package user

import (
	"aiscope/pkg/models/kubeconfig"
	"aiscope/pkg/utils/sliceutil"
	"golang.org/x/crypto/bcrypt"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"time"

	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	successSynced = "Synced"
	failedSynced  = "FailedSync"
	// is synced successfully
	messageResourceSynced = "User synced successfully"
	controllerName        = "user-controller"
	// user finalizer
	finalizer       = "finalizers.aiscope.io/users"
	interval        = time.Second
	timeout         = 15 * time.Second
	syncFailMessage = "Failed to sync: %s"
)

// Reconciler reconciles a User object
type Reconciler struct {
	client.Client
	Scheme                  *runtime.Scheme
	KubeconfigClient        kubeconfig.Interface
	Logger                  logr.Logger
	Recorder                record.EventRecorder
	MaxConcurrentReconciles int
}

//+kubebuilder:rbac:groups=iam.aiscope,resources=users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=iam.aiscope,resources=users/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=iam.aiscope,resources=users/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the User object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.WithValues("user", req.NamespacedName)
	user := &iamv1alpha2.User{}
	err := r.Get(ctx, req.NamespacedName, user)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if user.ObjectMeta.DeletionTimestamp.IsZero() {
		if !sliceutil.HasString(user.Finalizers, finalizer) {
			user.ObjectMeta.Finalizers = append(user.ObjectMeta.Finalizers, finalizer)
			if err = r.Update(ctx, user, &client.UpdateOptions{}); err != nil {
				logger.Error(err, "failed to update user")
				return ctrl.Result{}, err
			}
		}
	} else {
		if sliceutil.HasString(user.ObjectMeta.Finalizers, finalizer) {
			// TODO DELETE ALL resources of user

			// remove our finalizer from the list and update it.
			user.Finalizers = sliceutil.RemoveString(user.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})

			if err = r.Update(ctx, user, &client.UpdateOptions{}); err != nil {
				klog.Error(err)
				r.Recorder.Event(user, corev1.EventTypeWarning, failedSynced, fmt.Sprintf(syncFailMessage, err))
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, err
	}

	if err = r.encryptPassword(ctx, user); err != nil {
		klog.Error(err)
		r.Recorder.Event(user, corev1.EventTypeWarning, failedSynced, fmt.Sprintf(syncFailMessage, err))
		return ctrl.Result{}, err
	}
	if err = r.syncUserStatus(ctx, user); err != nil {
		klog.Error(err)
		r.Recorder.Event(user, corev1.EventTypeWarning, failedSynced, fmt.Sprintf(syncFailMessage, err))
		return ctrl.Result{}, err
	}

	if r.KubeconfigClient != nil {
		// ensure user KubeconfigClient configmap is created
		if err = r.KubeconfigClient.CreateKubeConfig(user); err != nil {
			klog.Error(err)
			r.Recorder.Event(user, corev1.EventTypeWarning, failedSynced, fmt.Sprintf(syncFailMessage, err))
			return ctrl.Result{}, err
		}
	}

	r.Recorder.Event(user, corev1.EventTypeNormal, successSynced, messageResourceSynced)

	return ctrl.Result{}, nil
}

// encryptPassword Encrypt and update the user password
func (r *Reconciler) encryptPassword(ctx context.Context, user *iamv1alpha2.User) error {
	// password is not empty and not encrypted
	if user.Spec.EncryptedPassword != "" && !isEncrypted(user.Spec.EncryptedPassword) {
		password, err := encrypt(user.Spec.EncryptedPassword)
		if err != nil {
			klog.Error(err)
			return err
		}
		user.Spec.EncryptedPassword = password
		if user.Annotations == nil {
			user.Annotations = make(map[string]string)
		}
		user.Annotations[iamv1alpha2.LastPasswordChangeTimeAnnotation] = time.Now().UTC().Format(time.RFC3339)
		// ensure plain text password won't be kept anywhere
		delete(user.Annotations, corev1.LastAppliedConfigAnnotation)
		err = r.Update(ctx, user, &client.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}


// syncUserStatus Update the user status
func (r *Reconciler) syncUserStatus(ctx context.Context, user *iamv1alpha2.User) error {
	if user.Spec.EncryptedPassword == "" {
		if user.Labels[iamv1alpha2.IdentifyProviderLabel] != "" {
			// mapped user from other identity provider always active until disabled
			if user.Status.State == nil || *user.Status.State != iamv1alpha2.UserActive {
				active := iamv1alpha2.UserActive
				user.Status = iamv1alpha2.UserStatus{
					State:              &active,
					LastTransitionTime: &metav1.Time{Time: time.Now()},
				}
				err := r.Update(ctx, user, &client.UpdateOptions{})
				if err != nil {
					return err
				}
			}
		} else {
			// becomes disabled after setting a blank password
			if user.Status.State == nil || *user.Status.State != iamv1alpha2.UserDisabled {
				disabled := iamv1alpha2.UserDisabled
				user.Status = iamv1alpha2.UserStatus{
					State:              &disabled,
					LastTransitionTime: &metav1.Time{Time: time.Now()},
				}
				err := r.Update(ctx, user, &client.UpdateOptions{})
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	// becomes active after password encrypted
	if isEncrypted(user.Spec.EncryptedPassword) {
		if user.Status.State == nil || *user.Status.State == iamv1alpha2.UserDisabled {
			active := iamv1alpha2.UserActive
			user.Status = iamv1alpha2.UserStatus{
				State:              &active,
				LastTransitionTime: &metav1.Time{Time: time.Now()},
			}
			err := r.Update(ctx, user, &client.UpdateOptions{})
			if err != nil {
				return err
			}
		}
	}

	// blocked user, check if need to unblock user
	if user.Status.State != nil && *user.Status.State == iamv1alpha2.UserAuthLimitExceeded {
		if user.Status.LastTransitionTime != nil {
			// && user.Status.LastTransitionTime.Add(r.AuthenticationOptions.AuthenticateRateLimiterDuration).Before(time.Now()) {
			// unblock user
			active := iamv1alpha2.UserActive
			user.Status = iamv1alpha2.UserStatus{
				State:              &active,
				LastTransitionTime: &metav1.Time{Time: time.Now()},
			}
			err := r.Update(ctx, user, &client.UpdateOptions{})
			if err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}

func encrypt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// isEncrypted returns whether the given password is encrypted
func isEncrypted(password string) bool {
	// bcrypt.Cost returns the hashing cost used to create the given hashed
	cost, _ := bcrypt.Cost([]byte(password))
	// cost > 0 means the password has been encrypted
	return cost > 0
}

// SetupWithManager sets up the controller with the Manager.
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
		For(&iamv1alpha2.User{}).
		Complete(r)
}
