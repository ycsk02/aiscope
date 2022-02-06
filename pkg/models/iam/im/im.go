package im

import (
	"aiscope/pkg/api"
	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
	"aiscope/pkg/apiserver/query"
	aiscope "aiscope/pkg/client/clientset/versioned"
	resources "aiscope/pkg/models/resources/v1alpha2"
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type IdentityManagementInterface interface {
	CreateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error)
	ListUsers(query *query.Query) (result *api.ListResult, err error)
	DescribeUser(username string) (*iamv1alpha2.User, error)
}

func NewOperator(aiClient aiscope.Interface) IdentityManagementInterface {
	im := &imOperator{
		aiClient:          aiClient,
	}
	return im
}

type imOperator struct {
	aiClient          aiscope.Interface
	userGetter        resources.Interface
	loginRecordGetter resources.Interface
}

func (im *imOperator) CreateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	user, err := im.aiClient.IamV1alpha2().Users().Create(context.Background(), user, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return user, nil
}

func (im *imOperator) ListUsers(query *query.Query) (result *api.ListResult, err error) {
	result, err = im.userGetter.List("", query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	items := make([]interface{}, 0)
	for _, item := range result.Items {
		user := item.(*iamv1alpha2.User)
		out := ensurePasswordNotOutput(user)
		items = append(items, out)
	}
	result.Items = items
	return result, nil
}


func (im *imOperator) DescribeUser(username string) (*iamv1alpha2.User, error) {
	obj, err := im.userGetter.Get("", username)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	user := obj.(*iamv1alpha2.User)
	return ensurePasswordNotOutput(user), nil
}

func ensurePasswordNotOutput(user *iamv1alpha2.User) *iamv1alpha2.User {
	out := user.DeepCopy()
	// ensure encrypted password will not be output
	out.Spec.EncryptedPassword = ""
	return out
}