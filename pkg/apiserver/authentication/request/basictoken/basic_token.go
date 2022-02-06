// Following code copied from k8s.io/apiserver/pkg/authorization/authorizerfactory to avoid import collision

package basictoken

import (
	"context"
	"errors"
	"net/http"

	"k8s.io/apiserver/pkg/authentication/authenticator"
)

type Password interface {
	AuthenticatePassword(ctx context.Context, user, password string) (*authenticator.Response, bool, error)
}

type Authenticator struct {
	auth Password
}

func New(auth Password) *Authenticator {
	return &Authenticator{auth}
}

var invalidToken = errors.New("invalid basic token")

func (a *Authenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {

	username, password, ok := req.BasicAuth()

	if !ok {
		return nil, false, nil
	}

	resp, ok, err := a.auth.AuthenticatePassword(req.Context(), username, password)

	// If the token authenticator didn't error, provide a default error
	if !ok && err == nil {
		err = invalidToken
	}

	return resp, ok, err
}
