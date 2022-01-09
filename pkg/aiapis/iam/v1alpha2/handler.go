package v1alpha2

import (
	"aiscope/pkg/api"
	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
	"aiscope/pkg/models/iam/im"
	"github.com/emicklei/go-restful"
)

type iamHandler struct {
	im         im.IdentityManagementInterface
}

func newIAMHandler(im im.IdentityManagementInterface) *iamHandler {
	return &iamHandler{
		im:         im,
	}
}

func (h *iamHandler) CreateUser(req *restful.Request, resp *restful.Response) {
	var user iamv1alpha2.User
	err := req.ReadEntity(&user)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	created, err := h.im.CreateUser(&user)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	// ensure encrypted password will not be output
	created.Spec.EncryptedPassword = ""

	resp.WriteEntity(created)
}


