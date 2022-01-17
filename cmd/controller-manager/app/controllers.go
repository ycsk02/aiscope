package app

import (
	"aiscope/pkg/authentication"
	"aiscope/pkg/controller/certificatesigningrequest"
	"aiscope/pkg/controller/clusterrolebinding"
	"aiscope/pkg/controller/globalrole"
	"aiscope/pkg/controller/globalrolebinding"
	"aiscope/pkg/controller/group"
	"aiscope/pkg/controller/groupbinding"
	"aiscope/pkg/controller/loginrecord"
	"aiscope/pkg/informers"
	"aiscope/pkg/simple/client/k8s"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func addControllers(mgr manager.Manager,
	client k8s.Client,
	informerFactory informers.InformerFactory,
	authenticationOptions *authentication.Options,
	kubectlImage string,
	stopCh <-chan struct{}) error {
	kubernetesInformer := informerFactory.KubernetesSharedInformerFactory()
	aiscopeInformer := informerFactory.AIScopeSharedInformerFactory()

	csrController := certificatesigningrequest.NewController(client.Kubernetes(),
		kubernetesInformer.Certificates().V1().CertificateSigningRequests(),
		kubernetesInformer.Core().V1().ConfigMaps(), client.Config())

	loginRecordController := loginrecord.NewLoginRecordController(
		client.Kubernetes(),
		client.AIScope(),
		aiscopeInformer.Iam().V1alpha2().LoginRecords(),
		aiscopeInformer.Iam().V1alpha2().Users(),
		authenticationOptions.LoginHistoryRetentionPeriod,
		authenticationOptions.LoginHistoryMaximumEntries)

	clusterRoleBindingController := clusterrolebinding.NewController(client.Kubernetes(),
		kubernetesInformer.Rbac().V1().ClusterRoleBindings(),
		kubernetesInformer.Apps().V1().Deployments(),
		kubernetesInformer.Core().V1().Pods(),
		aiscopeInformer.Iam().V1alpha2().Users(),
		kubectlImage)

	globalRoleController := globalrole.NewController(client.Kubernetes(), client.AIScope(),
		aiscopeInformer.Iam().V1alpha2().GlobalRoles())

	globalRoleBindingController := globalrolebinding.NewController(client.Kubernetes(), client.AIScope(),
		aiscopeInformer.Iam().V1alpha2().GlobalRoleBindings())

	groupBindingController := groupbinding.NewController(client.Kubernetes(), client.AIScope(),
		aiscopeInformer.Iam().V1alpha2().GroupBindings())

	groupController := group.NewController(client.Kubernetes(), client.AIScope(),
		aiscopeInformer.Iam().V1alpha2().Groups())

	controllers := map[string]manager.Runnable{
		"csr-controller":                csrController,
		"loginrecord-controller":        loginRecordController,
		"clusterrolebinding-controller": clusterRoleBindingController,
		"globalrolebinding-controller":  globalRoleBindingController,
		"groupbinding-controller":       groupBindingController,
		"group-controller":              groupController,
		"globalrole-controller":         globalRoleController,
	}

	for name, ctrl := range controllers {
		if ctrl == nil {
			klog.V(4).Infof("%s is not going to run due to dependent component disabled.", name)
			continue
		}

		if err := mgr.Add(ctrl); err != nil {
			klog.Error(err, "add controller to manager failed", "name", name)
			return err
		}
	}

	return nil
}
