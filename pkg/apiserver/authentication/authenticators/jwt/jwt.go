package jwt

import (
	"context"

	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"

	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"

	"aiscope/pkg/models/auth"

	iamv1alpha2listers "aiscope/pkg/client/listers/iam/v1alpha2"
)

// TokenAuthenticator implements kubernetes token authenticate interface with our custom logic.
// TokenAuthenticator will retrieve user info from cache by given token. If empty or invalid token
// was given, authenticator will still give passed response at the condition user will be user.Anonymous
// and group from user.AllUnauthenticated. This helps requests be passed along the handler chain,
// because some resources are public accessible.
type tokenAuthenticator struct {
	tokenOperator auth.TokenManagementInterface
	userLister    iamv1alpha2listers.UserLister
}

func NewTokenAuthenticator(tokenOperator auth.TokenManagementInterface, userLister iamv1alpha2listers.UserLister) authenticator.Token {
	return &tokenAuthenticator{
		tokenOperator: tokenOperator,
		userLister:    userLister,
	}
}

func (t *tokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	verified, err := t.tokenOperator.Verify(token)
	if err != nil {
		klog.Warning(err)
		return nil, false, err
	}

	if verified.User.GetName() == iamv1alpha2.PreRegistrationUser {
		return &authenticator.Response{
			User: verified.User,
		}, true, nil
	}

	u, err := t.userLister.Get(verified.User.GetName())
	if err != nil {
		return nil, false, err
	}

	return &authenticator.Response{
		User: &user.DefaultInfo{
			Name:   u.GetName(),
			Groups: append(u.Spec.Groups, user.AllAuthenticated),
		},
	}, true, nil
}
