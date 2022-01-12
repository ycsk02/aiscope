package v1alpha2

import (
	"aiscope/pkg/api"
	tenantv1alpha2 "aiscope/pkg/apis/tenant/v1alpha2"
	"aiscope/pkg/apiserver/runtime"
	aiscope "aiscope/pkg/client/clientset/versioned"
	"aiscope/pkg/constants"
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"net/http"
)

func AddToContainer(container *restful.Container, aiscope aiscope.Interface) error {
	ws := runtime.NewWebService(tenantv1alpha2.SchemeGroupVersion)
	handler := newTenantHandler(aiscope)

	ws.Route(ws.POST("/workspaces").
		To(handler.CreateWorkspace).
		Reads(tenantv1alpha2.Workspace{}).
		Returns(http.StatusOK, api.StatusOK, tenantv1alpha2.Workspace{}).
		Doc("Create workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceTag}))

	container.Add(ws)
	return nil
}
