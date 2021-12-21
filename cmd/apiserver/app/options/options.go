package options

import (
	"aiscope/pkg/apiserver"
	genericoptions "aiscope/pkg/server/options"
	"fmt"
	"net/http"
)

type ServerRunOptions struct {
	ConfigFile              string
	GenericServerRunOptions *genericoptions.ServerRunOptions
	DebugMode bool
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
	}

	return s
}

// NewAPIServer creates an APIServer instance using given options
func (s *ServerRunOptions) NewAPIServer(stopCh <-chan struct{}) (*apiserver.APIServer, error) {
	apiServer := &apiserver.APIServer{}

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", s.GenericServerRunOptions.InsecurePort),
	}

	apiServer.Server = server

	return apiServer, nil
}
