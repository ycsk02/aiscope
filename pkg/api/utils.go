package api

import (
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"net/http"
	"runtime"
	"strings"
)

var sanitizer = strings.NewReplacer(`&`, "&amp;", `<`, "&lt;", `>`, "&gt;")

// HandleBadRequest writes http.StatusBadRequest and log error
func HandleBadRequest(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusBadRequest, response, req, err)
}

func HandleError(response *restful.Response, req *restful.Request, err error) {
	var statusCode int
	switch t := err.(type) {
	case errors.APIStatus:
		statusCode = int(t.Status().Code)
	case restful.ServiceError:
		statusCode = t.Code
	default:
		statusCode = http.StatusInternalServerError
	}
	handle(statusCode, response, req, err)
}

func handle(statusCode int, response *restful.Response, req *restful.Request, err error) {
	_, fn, line, _ := runtime.Caller(2)
	klog.Errorf("%s:%d %v", fn, line, err)
	http.Error(response, sanitizer.Replace(err.Error()), statusCode)
}

func HandleNotFound(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusNotFound, response, req, err)
}

func HandleForbidden(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusForbidden, response, req, err)
}

func HandleInternalError(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusInternalServerError, response, req, err)
}
