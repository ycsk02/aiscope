package namespace

import (
	tenantv1alpha2 "aiscope/pkg/apis/tenant/v1alpha2"
	"aiscope/pkg/constants"
	controllerutils "aiscope/pkg/controller/utils/controller"
	"aiscope/pkg/utils/k8sutil"
	"aiscope/pkg/utils/sliceutil"
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	controllerName = "namespace-controller"
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
	logger := r.Logger.WithValues("namespace", req.NamespacedName)
	rootCtx := context.Background()
	namespace := &corev1.Namespace{}
	if err := r.Get(rootCtx, req.NamespacedName, namespace); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	finalizer := "finalizers.aiscope.io/namespace"

	if namespace.ObjectMeta.DeletionTimestamp.IsZero() {
		if !sliceutil.HasString(namespace.ObjectMeta.Finalizers, finalizer) {
			namespace.ObjectMeta.Finalizers = append(namespace.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(rootCtx, namespace); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if sliceutil.HasString(namespace.ObjectMeta.Finalizers, finalizer) {
			namespace.ObjectMeta.Finalizers = sliceutil.RemoveString(namespace.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})
			logger.V(4).Info("update namespace")
			if err := r.Update(rootCtx,namespace); err != nil {
				logger.Error(err, "update namespace failed")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	_, hasWorkspaceLabel := namespace.Labels[tenantv1alpha2.WorkspaceLabel]

	if hasWorkspaceLabel {
		if err := r.bindWorkspace(rootCtx, logger, namespace); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if err := r.unbindWorkspace(rootCtx, logger, namespace); err != nil {
			return ctrl.Result{}, err
		}
	}

	r.Recorder.Event(namespace, corev1.EventTypeNormal, controllerutils.SuccessSynced, controllerutils.MessageResourceSynced)
	return ctrl.Result{}, nil
}

func (r *Reconciler) bindWorkspace(ctx context.Context, logger logr.Logger, namespace *corev1.Namespace) error {
	workspace := &tenantv1alpha2.Workspace{}
	if err := r.Get(ctx, types.NamespacedName{Name: namespace.Labels[constants.WorkspaceLabelKey]}, workspace); err != nil {
		if errors.IsNotFound(err) && k8sutil.IsControlledBy(namespace.OwnerReferences, tenantv1alpha2.ResourceKindWorkspace, "") {
			return r.unbindWorkspace(ctx, logger, namespace)
		}

		return client.IgnoreNotFound(err)
	}

	if !workspace.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.unbindWorkspace(ctx, logger, namespace)
	}

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

func (r *Reconciler) unbindWorkspace(ctx context.Context, logger logr.Logger, namespace *corev1.Namespace) error {
	_, hasWorkspaceLabel := namespace.Labels[tenantv1alpha2.WorkspaceLabel]
	if hasWorkspaceLabel || k8sutil.IsControlledBy(namespace.OwnerReferences, tenantv1alpha2.ResourceKindWorkspace, "") {
		ns := namespace.DeepCopy()

		wsName := k8sutil.GetWorkspaceOwnerName(ns.OwnerReferences)
		if hasWorkspaceLabel {
			wsName = namespace.Labels[tenantv1alpha2.WorkspaceLabel]
		}

		delete(ns.Labels, constants.WorkspaceLabelKey)
		ns.OwnerReferences = k8sutil.RemoveWorkspaceOwnerReference(ns.OwnerReferences)
		logger.V(4).Info("remove owner reference and label", "namespace", ns.Name, "workspace", wsName)
		if err := r.Update(ctx, ns); err != nil {
			logger.Error(err, "update owner reference failed")
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
		For(&corev1.Namespace{}).
		Complete(r)
}

