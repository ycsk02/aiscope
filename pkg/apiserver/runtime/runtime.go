package runtime

import (
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ApiRootPath = "/aiapis"
)

func NewWebService(gv schema.GroupVersion) *restful.WebService {
	webservice := restful.WebService{}
	webservice.Path(ApiRootPath + "/" + gv.String()).
		Produces(restful.MIME_JSON)

	return &webservice
}
