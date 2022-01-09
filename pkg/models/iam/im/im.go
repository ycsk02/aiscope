package im

import (
	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
	aiscope "aiscope/pkg/client/clientset/versioned"
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type IdentityManagementInterface interface {
	CreateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error)
}

func NewOperator(aiClient aiscope.Interface) IdentityManagementInterface {
	im := &imOperator{
		aiClient:          aiClient,
	}
	return im
}

type imOperator struct {
	aiClient          aiscope.Interface
}

func (im *imOperator) CreateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	user, err := im.aiClient.IamV1alpha2().Users().Create(context.Background(), user, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return user, nil
}