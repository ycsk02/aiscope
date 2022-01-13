package apiserver

import (
	iamapi "aiscope/pkg/aiapis/iam/v1alpha2"
	tenantapi "aiscope/pkg/aiapis/tenant/v1alpha2"
	"aiscope/pkg/aiapis/version"
	apiserverconfig "aiscope/pkg/apiserver/config"
	"aiscope/pkg/apiserver/filters"
	"aiscope/pkg/apiserver/request"
	"aiscope/pkg/models/iam/im"
	"aiscope/pkg/simple/client/k8s"
	"context"
	"github.com/emicklei/go-restful"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog/v2"
	"net/http"
)

type APIServer struct {
	ServerCount int

	Server *http.Server

	Config *apiserverconfig.Config

	// webservice container, where all webservice defines
	container *restful.Container

	KubernetesClient k8s.Client
}

func (s *APIServer) PrepareRun(stopCh <-chan struct{}) error {
	s.container = restful.NewContainer()

	s.container.Router(restful.CurlyRouter{})

	for _, ws := range s.container.RegisteredWebServices() {
		klog.V(2).Infof("%s", ws.RootPath())
	}

	s.Server.Handler = s.container

	s.installAIscopeAPIs()

	s.buildHandlerChain(stopCh)

	return nil
}

func (s *APIServer) installAIscopeAPIs() {
	imOperator := im.NewOperator(s.KubernetesClient.AIScope())

	urlruntime.Must(version.AddToContainer(s.container, s.KubernetesClient.Discovery()))

	urlruntime.Must(iamapi.AddToContainer(s.container, imOperator))

	urlruntime.Must(tenantapi.AddToContainer(s.container, s.KubernetesClient.AIScope(), s.KubernetesClient.Kubernetes()))
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

func (s *APIServer) buildHandlerChain(stopCh <-chan struct{}) {
	requestInfoResolver := &request.RequestInfoFactory{
		APIPrefixes:          sets.NewString("api", "apis"),
	}

	handler := s.Server.Handler
	handler = filters.WithKubeAPIServer(handler, s.KubernetesClient.Config(), &errorResponder{})

	handler = filters.WithRequestInfo(handler, requestInfoResolver)

	s.Server.Handler = handler
}

type errorResponder struct{}

func (e *errorResponder) Error(w http.ResponseWriter, req *http.Request, err error) {
	klog.Error(err)
	responsewriters.InternalError(w, req, err)
}

