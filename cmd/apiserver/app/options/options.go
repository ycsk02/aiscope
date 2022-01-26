package options

import (
	"aiscope/pkg/apiserver"
	apiserverconfig "aiscope/pkg/apiserver/config"
	"aiscope/pkg/informers"
	genericoptions "aiscope/pkg/server/options"
	"aiscope/pkg/simple/client/k8s"
	"fmt"
	"net/http"
)

type ServerRunOptions struct {
	ConfigFile              string
	GenericServerRunOptions *genericoptions.ServerRunOptions
	*apiserverconfig.Config

	DebugMode bool
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		GenericServerRunOptions:    genericoptions.NewServerRunOptions(),
		Config:                     apiserverconfig.New(),
	}

	return s
}

// NewAPIServer creates an APIServer instance using given options
func (s *ServerRunOptions) NewAPIServer(stopCh <-chan struct{}) (*apiserver.APIServer, error) {
	apiServer := &apiserver.APIServer{
		Config:     s.Config,
	}

	kubernetesClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		return nil, err
	}
	apiServer.KubernetesClient = kubernetesClient

	informerFactory := informers.NewInformerFactories(kubernetesClient.Kubernetes(), kubernetesClient.AIScope())
	apiServer.InformerFactory = informerFactory

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", s.GenericServerRunOptions.InsecurePort),
	}

	apiServer.Server = server

	return apiServer, nil
}
