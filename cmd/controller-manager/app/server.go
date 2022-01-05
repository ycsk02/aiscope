package app

import (
	"aiscope/cmd/controller-manager/app/options"
	"aiscope/pkg/apis"
	"aiscope/pkg/controller/namespace"
	"aiscope/pkg/controller/user"
	"aiscope/pkg/controller/workspace"
	"aiscope/pkg/informers"
	"aiscope/pkg/models/kubeconfig"
	"aiscope/pkg/simple/client/k8s"
	"context"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func NewControllerManagerCommand() *cobra.Command {

	s := options.NewAIScopeControllerManagerOptions()

	cmd := &cobra.Command{
		Use: "controller-manager",
		Long: `AIScope controller manager`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(s, signals.SetupSignalHandler()); err != nil {
				klog.Error(err)
				os.Exit(1)
			}
		},
		SilenceUsage: true,
	}

	return cmd
}

func run(s *options.AIScopeControllerManagerOptions, ctx context.Context) error {
	kubernetesClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		klog.Errorf("Failed to create kubernetes clientset %v", err)
		return err
	}

	informerFactory := informers.NewInformerFactories(
		kubernetesClient.Kubernetes())

	mgrOptions := manager.Options{
		Port: 8443,
	}

	if s.LeaderElect {
		mgrOptions = manager.Options{
			Port:                    8443,
			LeaderElection:          s.LeaderElect,
			LeaderElectionNamespace: "aiscope-system",
			LeaderElectionID:        "aiscope-controller-manager-leader-election",
			LeaseDuration:           &s.LeaderElection.LeaseDuration,
			RetryPeriod:             &s.LeaderElection.RetryPeriod,
			RenewDeadline:           &s.LeaderElection.RenewDeadline,
		}
	}

	klog.V(0).Info("setting up manager")
	ctrl.SetLogger(klogr.New())

	mgr, err := manager.New(kubernetesClient.Config(), mgrOptions)
	if err != nil {
		klog.Fatalf("unable to set up overall controller manager: %v", err)
	}

	if err = apis.AddToScheme(mgr.GetScheme()); err != nil {
		klog.Fatalf("unable add APIs to scheme: %v", err)
	}

	// register common meta types into schemas.
	metav1.AddToGroupVersion(mgr.GetScheme(), metav1.SchemeGroupVersion)

	workspaceReconciler := &workspace.Reconciler{}
	if err = workspaceReconciler.SetupWithManager(mgr); err != nil {
		klog.Fatalf("Unable to create workspace controller: %v", err)
	}

	namespaeReconciler := &namespace.Reconciler{}
	if err = namespaeReconciler.SetupWithManager(mgr); err != nil {
		klog.Fatalf("Unable to create namespace controller: %v", err)
	}

	kubeconfigClient := kubeconfig.NewOperator(kubernetesClient.Kubernetes(),
		informerFactory.KubernetesSharedInformerFactory().Core().V1().ConfigMaps().Lister(),
		kubernetesClient.Config())
	userController := user.Reconciler{
		MaxConcurrentReconciles: 4,
		KubeconfigClient:        kubeconfigClient,
	}
	if err = userController.SetupWithManager(mgr); err != nil {
		klog.Fatalf("Unable to create user controller: %v", err)
	}

	// Start cache data after all informer is registered
	klog.V(0).Info("Starting cache resource from apiserver...")
	informerFactory.Start(ctx.Done())

	klog.V(0).Info("Starting the controllers.")
	if err = mgr.Start(ctx); err != nil {
		klog.Fatalf("unable to run the manager: %v", err)
	}

	return nil
}
