package app

import (
	"aiscope/pkg/controller/certificatesigningrequest"
	"aiscope/pkg/informers"
	"aiscope/pkg/simple/client/k8s"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func addControllers(mgr manager.Manager, client k8s.Client, informerFactory informers.InformerFactory, stopCh <-chan struct{}) error {
	kubernetesInformer := informerFactory.KubernetesSharedInformerFactory()

	csrController := certificatesigningrequest.NewController(client.Kubernetes(),
		kubernetesInformer.Certificates().V1().CertificateSigningRequests(),
		kubernetesInformer.Core().V1().ConfigMaps(), client.Config())

	controllers := map[string]manager.Runnable{
		"csr-controller":                csrController,
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
