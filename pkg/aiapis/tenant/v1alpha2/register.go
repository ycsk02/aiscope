package v1alpha2

import (
	"aiscope/pkg/api"
	tenantv1alpha2 "aiscope/pkg/apis/tenant/v1alpha2"
	"aiscope/pkg/apiserver/runtime"
	aiscope "aiscope/pkg/client/clientset/versioned"
	"aiscope/pkg/constants"
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

func AddToContainer(container *restful.Container, aiscope aiscope.Interface, k8sclient kubernetes.Interface) error {
	ws := runtime.NewWebService(tenantv1alpha2.SchemeGroupVersion)
	handler := newTenantHandler(aiscope, k8sclient)

	ws.Route(ws.POST("/workspaces").
		To(handler.CreateWorkspace).
		Reads(tenantv1alpha2.Workspace{}).
		Returns(http.StatusOK, api.StatusOK, tenantv1alpha2.Workspace{}).
		Doc("Create workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceTag}))

	ws.Route(ws.POST("/workspaces/{workspace}/namespaces").
		To(handler.CreateNamespace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("Create the namespace of the specified workspace for the current user").
		Reads(corev1.Namespace{}).
		Returns(http.StatusOK, api.StatusOK, corev1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/namespaces").
		To(handler.ListNamespaces).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List the namespaces of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/workspacemembers/{workspacemember}/namespaces").
		To(handler.ListNamespaces).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacemember", "workspacemember username")).
		Doc("List the namespaces of the specified workspace for the workspace member").
		Reads(corev1.Namespace{}).
		Returns(http.StatusOK, api.StatusOK, corev1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceTag}))

	container.Add(ws)
	return nil
}
