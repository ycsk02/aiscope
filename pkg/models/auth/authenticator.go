package auth

import (
	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
	"aiscope/pkg/apiserver/authentication/identityprovider"
	iamv1alpha2listers "aiscope/pkg/client/listers/iam/v1alpha2"
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	authuser "k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"
	"net/http"
	"net/mail"
	"strings"
)

// PasswordAuthenticator is an interface implemented by authenticator which take a
// username and password.
type PasswordAuthenticator interface {
	Authenticate(ctx context.Context, username, password string) (authuser.Info, string, error)
}

type OAuthAuthenticator interface {
	Authenticate(ctx context.Context, provider string, req *http.Request) (authuser.Info, string, error)
}

type userGetter struct {
	userLister iamv1alpha2listers.UserLister
}

func preRegistrationUser(idp string, identity identityprovider.Identity) authuser.Info {
	return &authuser.DefaultInfo{
		Name: iamv1alpha2.PreRegistrationUser,
		Extra: map[string][]string{
			iamv1alpha2.ExtraIdentityProvider: {idp},
			iamv1alpha2.ExtraUID:              {identity.GetUserID()},
			iamv1alpha2.ExtraUsername:         {identity.GetUsername()},
			iamv1alpha2.ExtraEmail:            {identity.GetEmail()},
		},
	}
}

func mappedUser(idp string, identity identityprovider.Identity) *iamv1alpha2.User {
	// username convert
	username := strings.ToLower(identity.GetUsername())
	return &iamv1alpha2.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: username,
			Labels: map[string]string{
				iamv1alpha2.IdentifyProviderLabel: idp,
				iamv1alpha2.OriginUIDLabel:        identity.GetUserID(),
			},
		},
		Spec: iamv1alpha2.UserSpec{Email: identity.GetEmail()},
	}
}

// findUser returns the user associated with the username or email
func (u *userGetter) findUser(username string) (*iamv1alpha2.User, error) {
	if _, err := mail.ParseAddress(username); err != nil {
		return u.userLister.Get(username)
	}

	users, err := u.userLister.List(labels.Everything())
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, user := range users {
		if user.Spec.Email == username {
			return user, nil
		}
	}

	return nil, errors.NewNotFound(iamv1alpha2.Resource("user"), username)
}

// findMappedUser returns the user which mapped to the identity
func (u *userGetter) findMappedUser(idp, uid string) (*iamv1alpha2.User, error) {
	selector := labels.SelectorFromSet(labels.Set{
		iamv1alpha2.IdentifyProviderLabel: idp,
		// iamv1alpha2.OriginUIDLabel:        uid,
	})

	users, err := u.userLister.List(selector)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	if len(users) != 1 {
		return nil, errors.NewNotFound(iamv1alpha2.Resource("user"), uid)
	}

	return users[0], err
}
