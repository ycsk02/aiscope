package runtime

import (
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ApiRootPath = "/aiapis"
)

const MimeMergePatchJson = "application/merge-patch+json"
const MimeJsonPatchJson = "application/json-patch+json"
const MimeMultipartFormData = "multipart/form-data"

func NewWebService(gv schema.GroupVersion) *restful.WebService {
	webservice := restful.WebService{}
	webservice.Path(ApiRootPath + "/" + gv.String()).
		Produces(restful.MIME_JSON)

	return &webservice
}
