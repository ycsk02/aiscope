package ldap

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"

	"aiscope/pkg/api"
	"aiscope/pkg/apiserver/query"
)

const FAKE_HOST string = "FAKE"

// simpleLdap is a implementation of ldap.Interface, you should never use this in production env!
type simpleLdap struct {
	store map[string]*iamv1alpha2.User
}

func NewSimpleLdap() Interface {
	sl := &simpleLdap{
		store: map[string]*iamv1alpha2.User{},
	}

	// initialize with a admin user
	admin := &iamv1alpha2.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "admin",
		},
		Spec: iamv1alpha2.UserSpec{
			Email:             "admin@aiscope.io",
			Lang:              "eng",
			Description:       "administrator",
			Groups:            nil,
			EncryptedPassword: "P@88w0rd",
		},
	}
	sl.store[admin.Name] = admin
	return sl
}

func (s simpleLdap) Create(user *iamv1alpha2.User) error {
	s.store[user.Name] = user
	return nil
}

func (s simpleLdap) Update(user *iamv1alpha2.User) error {
	_, err := s.Get(user.Name)
	if err != nil {
		return err
	}
	s.store[user.Name] = user
	return nil
}

func (s simpleLdap) Delete(name string) error {
	_, err := s.Get(name)
	if err != nil {
		return err
	}
	delete(s.store, name)
	return nil
}

func (s simpleLdap) Get(name string) (*iamv1alpha2.User, error) {
	if user, ok := s.store[name]; !ok {
		return nil, ErrUserNotExists
	} else {
		return user, nil
	}
}

func (s simpleLdap) Authenticate(name string, password string) error {
	if user, err := s.Get(name); err != nil {
		return err
	} else {
		if user.Spec.EncryptedPassword != password {
			return ErrInvalidCredentials
		}
	}

	return nil
}

func (l *simpleLdap) List(query *query.Query) (*api.ListResult, error) {
	items := make([]interface{}, 0)

	for _, user := range l.store {
		items = append(items, user)
	}

	return &api.ListResult{
		Items:      items,
		TotalItems: len(items),
	}, nil
}
