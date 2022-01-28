package options

import (
	"aiscope/pkg/apiserver"
	apiserverconfig "aiscope/pkg/apiserver/config"
	"aiscope/pkg/informers"
	genericoptions "aiscope/pkg/server/options"
	"aiscope/pkg/simple/client/cache"
	"aiscope/pkg/simple/client/k8s"
	"fmt"
	"k8s.io/klog/v2"
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

const fakeInterface string = "FAKE"

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

	var cacheClient cache.Interface
	if s.RedisOptions != nil && len(s.RedisOptions.Host) != 0 {
		if s.RedisOptions.Host == fakeInterface && s.DebugMode {
			apiServer.CacheClient = cache.NewSimpleCache()
		} else {
			cacheClient, err = cache.NewRedisClient(s.RedisOptions, stopCh)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to redis service, please check redis status, error: %v", err)
			}
			apiServer.CacheClient = cacheClient
		}
	} else {
		klog.Warning("apiserver starts without redis provided, it will use in memory cache. " +
			"This may cause inconsistencies when running apiserver with multiple replicas.")
		apiServer.CacheClient = cache.NewSimpleCache()
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", s.GenericServerRunOptions.InsecurePort),
	}

	apiServer.Server = server

	return apiServer, nil
}
