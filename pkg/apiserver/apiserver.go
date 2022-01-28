package apiserver

import (
	experimentapi "aiscope/pkg/aiapis/experiment/v1alpha2"
	iamapi "aiscope/pkg/aiapis/iam/v1alpha2"
	"aiscope/pkg/aiapis/oauth"
	tenantapi "aiscope/pkg/aiapis/tenant/v1alpha2"
	"aiscope/pkg/aiapis/version"
	"aiscope/pkg/apiserver/authentication/token"
	apiserverconfig "aiscope/pkg/apiserver/config"
	"aiscope/pkg/apiserver/filters"
	"aiscope/pkg/apiserver/request"
	"aiscope/pkg/informers"
	"aiscope/pkg/models/auth"
	"aiscope/pkg/models/experiment"
	"aiscope/pkg/models/iam/im"
	"aiscope/pkg/simple/client/cache"
	"aiscope/pkg/simple/client/k8s"
	"context"
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

	CacheClient cache.Interface

	// webservice container, where all webservice defines
	container *restful.Container

	KubernetesClient k8s.Client

	// entity that issues tokens
	Issuer token.Issuer

	InformerFactory informers.InformerFactory
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
	epOperator := experiment.New(s.KubernetesClient.AIScope(), s.InformerFactory)

	urlruntime.Must(version.AddToContainer(s.container, s.KubernetesClient.Discovery()))

	urlruntime.Must(iamapi.AddToContainer(s.container, imOperator))
	urlruntime.Must(experimentapi.AddToContainer(s.container, epOperator))
	urlruntime.Must(tenantapi.AddToContainer(s.container, s.KubernetesClient.AIScope(), s.KubernetesClient.Kubernetes()))

	userLister := s.InformerFactory.AIScopeSharedInformerFactory().Iam().V1alpha2().Users().Lister()
	urlruntime.Must(oauth.AddToContainer(s.container, imOperator,
		auth.NewTokenOperator(s.CacheClient, s.Issuer, s.Config.AuthenticationOptions),
		auth.NewOAuthAuthenticator(s.KubernetesClient.AIScope(), userLister, s.Config.AuthenticationOptions),
		auth.NewLoginRecorder(s.KubernetesClient.AIScope(), userLister),
		s.Config.AuthenticationOptions))
}

func (s *APIServer) Run(ctx context.Context) (err error) {

	err = s.waitForResourceSync(ctx)
	if err != nil {
		return err
	}

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

func (s *APIServer) waitForResourceSync(ctx context.Context) error {
	klog.V(0).Info("Start cache objects")

	stopCh := ctx.Done()

	discoveryClient := s.KubernetesClient.Kubernetes().Discovery()
	_, apiResourcesList, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return err
	}

	isResourceExists := func(resource schema.GroupVersionResource) bool {
		for _, apiResource := range apiResourcesList {
			if apiResource.GroupVersion == resource.GroupVersion().String() {
				for _, rsc := range apiResource.APIResources {
					if rsc.Name == resource.Resource {
						return true
					}
				}
			}
		}
		return false
	}

	aiGVRs := []schema.GroupVersionResource{
		{Group: "tenant.aiscope", Version: "v1alpha2", Resource: "workspaces"},
		{Group: "iam.aiscope", Version: "v1alpha2", Resource: "users"},
		{Group: "experiment.aiscope", Version: "v1alpha2", Resource: "jupyternotebooks"},
		{Group: "experiment.aiscope", Version: "v1alpha2", Resource: "trackingservers"},
	}

	aiInformerFactory := s.InformerFactory.AIScopeSharedInformerFactory()

	for _, gvr := range aiGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err = aiInformerFactory.ForResource(gvr)
			if err != nil {
				return err
			}
		}
	}
	aiInformerFactory.Start(stopCh)
	aiInformerFactory.WaitForCacheSync(stopCh)

	klog.V(0).Info("Finished caching objects")

	return nil
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

