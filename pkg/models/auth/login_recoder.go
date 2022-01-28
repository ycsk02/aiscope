package auth

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"

	aiscope "aiscope/pkg/client/clientset/versioned"
	iamv1alpha2listers "aiscope/pkg/client/listers/iam/v1alpha2"
)

type LoginRecorder interface {
	RecordLogin(username string, loginType iamv1alpha2.LoginType, provider string, sourceIP string, userAgent string, authErr error) error
}

type loginRecorder struct {
	aiClient   aiscope.Interface
	userGetter *userGetter
}

func NewLoginRecorder(aiClient aiscope.Interface, userLister iamv1alpha2listers.UserLister) LoginRecorder {
	return &loginRecorder{
		aiClient:   aiClient,
		userGetter: &userGetter{userLister: userLister},
	}
}

// RecordLogin Create v1alpha2.LoginRecord for existing accounts
func (l *loginRecorder) RecordLogin(username string, loginType iamv1alpha2.LoginType, provider, sourceIP, userAgent string, authErr error) error {
	// only for existing accounts, solve the problem of huge entries
	user, err := l.userGetter.findUser(username)
	if err != nil {
		// ignore not found error
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Error(err)
		return err
	}
	loginEntry := &iamv1alpha2.LoginRecord{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", user.Name),
			Labels: map[string]string{
				iamv1alpha2.UserReferenceLabel: user.Name,
			},
		},
		Spec: iamv1alpha2.LoginRecordSpec{
			Type:      loginType,
			Provider:  provider,
			Success:   true,
			Reason:    iamv1alpha2.AuthenticatedSuccessfully,
			SourceIP:  sourceIP,
			UserAgent: userAgent,
		},
	}

	if authErr != nil {
		loginEntry.Spec.Success = false
		loginEntry.Spec.Reason = authErr.Error()
	}

	_, err = l.aiClient.IamV1alpha2().LoginRecords().Create(context.Background(), loginEntry, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}
