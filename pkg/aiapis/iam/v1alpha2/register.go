package v1alpha2

import (
	"aiscope/pkg/api"
	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
	"aiscope/pkg/apiserver/runtime"
	"aiscope/pkg/constants"
	"aiscope/pkg/models/iam/im"
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"net/http"
)

func AddToContainer(container *restful.Container, im im.IdentityManagementInterface) error {
	ws := runtime.NewWebService(iamv1alpha2.SchemeGroupVersion)
	handler := newIAMHandler(im)

	// users
	ws.Route(ws.POST("/users").
		To(handler.CreateUser).
		Doc("Create a global user account.").
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Reads(iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))

	ws.Route(ws.GET("/users").
		To(handler.ListUsers).
		Doc("List all users.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.User{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))

	container.Add(ws)
	return nil
}
