package tenant

import (
	"aiscope/pkg/api"
	tenantv1alpha2 "aiscope/pkg/apis/tenant/v1alpha2"
	"aiscope/pkg/apiserver/query"
	"aiscope/pkg/apiserver/request"
	aiscope "aiscope/pkg/client/clientset/versioned"
	resources "aiscope/pkg/models/resources/v1alpha2"
	resourcev1alpha2 "aiscope/pkg/models/resources/v1alpha2/resource"
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type Interface interface {
	CreateWorkspace(workspace *tenantv1alpha2.Workspace) (*tenantv1alpha2.Workspace, error)
	CreateNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error)
	ListNamespaces(user user.Info, workspace string, queryParam *query.Query) (*api.ListResult, error)
}

type tenantOperator struct {
	aiClient            aiscope.Interface
	k8sclient           kubernetes.Interface
	resourceGetter      *resourcev1alpha2.ResourceGetter
}

func NewOperator(aiClient aiscope.Interface, k8sclient kubernetes.Interface) Interface {
	return &tenantOperator{
		aiClient:           aiClient,
		k8sclient:          k8sclient,
	}
}

func (t *tenantOperator) CreateWorkspace(workspace *tenantv1alpha2.Workspace) (*tenantv1alpha2.Workspace, error) {
	return t.aiClient.TenantV1alpha2().Workspaces().Create(context.Background(), workspace, metav1.CreateOptions{})
}

func (t *tenantOperator) CreateNamespace(workspace string, namespace *corev1.Namespace) (*corev1.Namespace, error) {
	return t.k8sclient.CoreV1().Namespaces().Create(context.Background(), labelNamespaceWithWorkspaceName(namespace, workspace), metav1.CreateOptions{})
}

func labelNamespaceWithWorkspaceName(namespace *corev1.Namespace, workspaceName string) *corev1.Namespace {
	if namespace.Labels == nil {
		namespace.Labels = make(map[string]string, 0)
	}

	namespace.Labels[tenantv1alpha2.WorkspaceLabel] = workspaceName // label namespace with workspace name

	return namespace
}

func (t *tenantOperator) ListNamespaces(user user.Info, workspace string, queryParam *query.Query) (*api.ListResult, error) {
	nsScope := request.ClusterScope
	if workspace != "" {
		nsScope = request.WorkspaceScope
		// filter by workspace
		queryParam.Filters[query.FieldLabel] = query.Value(fmt.Sprintf("%s=%s", tenantv1alpha2.WorkspaceLabel, workspace))
	}

	listNS := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		Workspace:       workspace,
		Resource:        "namespaces",
		ResourceRequest: true,
		ResourceScope:   nsScope,
	}

	decision, _, err := t.authorizer.Authorize(listNS)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// allowed to list all namespaces in the specified scope
	if decision == authorizer.DecisionAllow {
		result, err := t.resourceGetter.List("namespaces", "", queryParam)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		return result, nil
	}

	// retrieving associated resources through role binding
	roleBindings, err := t.am.ListRoleBindings(user.GetName(), user.GetGroups(), "")
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	namespaces := make([]runtime.Object, 0)
	for _, roleBinding := range roleBindings {
		obj, err := t.resourceGetter.Get("namespaces", "", roleBinding.Namespace)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		namespace := obj.(*corev1.Namespace)
		// label matching selector, remove duplicate entity
		if queryParam.Selector().Matches(labels.Set(namespace.Labels)) &&
			!contains(namespaces, namespace) {
			namespaces = append(namespaces, namespace)
		}
	}

	// use default pagination search logic
	result := resources.DefaultList(namespaces, queryParam, func(left runtime.Object, right runtime.Object, field query.Field) bool {
		return resources.DefaultObjectMetaCompare(left.(*corev1.Namespace).ObjectMeta, right.(*corev1.Namespace).ObjectMeta, field)
	}, func(object runtime.Object, filter query.Filter) bool {
		return resources.DefaultObjectMetaFilter(object.(*corev1.Namespace).ObjectMeta, filter)
	})

	return result, nil
}

func contains(objects []runtime.Object, object runtime.Object) bool {
	for _, item := range objects {
		if item == object {
			return true
		}
	}
	return false
}
