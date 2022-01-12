package tenant

import (
	tenantv1alpha2 "aiscope/pkg/apis/tenant/v1alpha2"
	aiscope "aiscope/pkg/client/clientset/versioned"
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Interface interface {
	CreateWorkspace(workspace *tenantv1alpha2.Workspace) (*tenantv1alpha2.Workspace, error)
}

type tenantOperator struct {
	aiClient          aiscope.Interface
}

func NewOperator(aiClient aiscope.Interface) Interface {
	return &tenantOperator{
		aiClient:          aiClient,
	}
}

func (t *tenantOperator) CreateWorkspace(workspace *tenantv1alpha2.Workspace) (*tenantv1alpha2.Workspace, error) {
	return t.aiClient.TenantV1alpha2().Workspaces().Create(context.Background(), workspace, metav1.CreateOptions{})
}
