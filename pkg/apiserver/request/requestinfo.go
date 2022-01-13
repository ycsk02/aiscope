package request

import (
	"aiscope/pkg/api"
	"aiscope/pkg/utils/iputil"
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/sets"
	k8srequest "k8s.io/apiserver/pkg/endpoints/request"
	"net/http"
	"strings"
)

const (
	GlobalScope             = "Global"
	ClusterScope            = "Cluster"
	WorkspaceScope          = "Workspace"
	NamespaceScope          = "Namespace"
)

type RequestInfoResolver interface {
	NewRequestInfo(req *http.Request) (*RequestInfo, error)
}

var specialVerbs = sets.NewString("proxy", "watch")

// RequestInfo holds information parsed from the http.Request,
// extended from k8s.io/apiserver/pkg/endpoints/request/requestinfo.go
type RequestInfo struct {
	*k8srequest.RequestInfo

	// IsKubernetesRequest indicates whether or not the request should be handled by kubernetes or kubesphere
	IsKubernetesRequest bool

	// Workspace of requested resource, for non-workspaced resources, this may be empty
	Workspace string

	// Cluster of requested resource, this is empty in single-cluster environment
	Cluster string

	// DevOps project of requested resource
	DevOps string

	// Scope of requested resource.
	ResourceScope string

	// Source IP
	SourceIP string

	// User agent
	UserAgent string
}

type RequestInfoFactory struct {
	APIPrefixes          sets.String
}

var kubernetesAPIPrefixes = sets.NewString("api", "apis")


func (r *RequestInfoFactory) NewRequestInfo(req *http.Request) (*RequestInfo, error) {
	requestInfo := RequestInfo{
		IsKubernetesRequest: false,
		RequestInfo: &k8srequest.RequestInfo{
			Path: req.URL.Path,
			Verb: req.Method,
		},
		Workspace: api.WorkspaceNone,
		Cluster:   api.ClusterNone,
		SourceIP:  iputil.RemoteIp(req),
		UserAgent: req.UserAgent(),
	}

	defer func() {
		prefix := requestInfo.APIPrefix
		if prefix == "" {
			currentParts := splitPath(requestInfo.Path)
			// Proxy discovery API
			if len(currentParts) > 0 && len(currentParts) < 3 {
				prefix = currentParts[0]
			}
		}
		if kubernetesAPIPrefixes.Has(prefix) {
			requestInfo.IsKubernetesRequest = true
		}
	}()

	currentParts := splitPath(req.URL.Path)
	if len(currentParts) < 3 {
		return &requestInfo, nil
	}

	if !r.APIPrefixes.Has(currentParts[0]) {
		// return a non-resource request
		return &requestInfo, nil
	}
	requestInfo.APIPrefix = currentParts[0]
	currentParts = currentParts[1:]

	requestInfo.IsResourceRequest = true
	requestInfo.APIVersion = currentParts[0]
	currentParts = currentParts[1:]

	if len(currentParts) > 0 && specialVerbs.Has(currentParts[0]) {
		if len(currentParts) < 2 {
			return &requestInfo, fmt.Errorf("unable to determine kind and namespace from url: %v", req.URL)
		}

		requestInfo.Verb = currentParts[0]
		currentParts = currentParts[1:]
	} else {
		switch req.Method {
		case "POST":
			requestInfo.Verb = "create"
		case "GET", "HEAD":
			requestInfo.Verb = "get"
		case "PUT":
			requestInfo.Verb = "update"
		case "PATCH":
			requestInfo.Verb = "patch"
		case "DELETE":
			requestInfo.Verb = "delete"
		default:
			requestInfo.Verb = ""
		}
	}

	return &requestInfo, nil
}

type requestInfoKeyType int

// requestInfoKey is the RequestInfo key for the context. It's of private type here. Because
// keys are interfaces and interfaces are equal when the type and the value is equal, this
// does not conflict with the keys defined in pkg/api.
const requestInfoKey requestInfoKeyType = iota

func WithRequestInfo(parent context.Context, info *RequestInfo) context.Context {
	return k8srequest.WithValue(parent, requestInfoKey, info)
}

func RequestInfoFrom(ctx context.Context) (*RequestInfo, bool) {
	info, ok := ctx.Value(requestInfoKey).(*RequestInfo)
	return info, ok
}

// splitPath returns the segments for a URL path.
func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}
