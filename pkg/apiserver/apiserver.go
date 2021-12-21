package apiserver

import (
	"context"
	"github.com/emicklei/go-restful"
	"k8s.io/klog/v2"
	"net/http"
)

type APIServer struct {
	// number of kubesphere apiserver
	ServerCount int

	Server *http.Server

	// webservice container, where all webservice defines
	container *restful.Container
}

func (s *APIServer) PrepareRun(stopCh <-chan struct{}) error {
	s.container = restful.NewContainer()

	s.container.Router(restful.CurlyRouter{})

	for _, ws := range s.container.RegisteredWebServices() {
		klog.V(2).Infof("%s", ws.RootPath())
	}

	s.Server.Handler = s.container

	return nil
}

func (s *APIServer) Run(ctx context.Context) (err error) {

	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-ctx.Done()
		_ = s.Server.Shutdown(shutdownCtx)
	}()

	klog.V(0).Infof("Start listening on %s", s.Server.Addr)
	if s.Server.TLSConfig != nil {
		err = s.Server.ListenAndServeTLS("", "")
	} else {
		err = s.Server.ListenAndServe()
	}

	return err
}

