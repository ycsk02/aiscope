package user

import (
	"aiscope/pkg/api"
	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
	tenantv1alpha2 "aiscope/pkg/apis/tenant/v1alpha2"
	"aiscope/pkg/apiserver/query"
	informers "aiscope/pkg/client/informers/externalversions"
	"aiscope/pkg/models/resources/v1alpha2"
	"aiscope/pkg/utils/sliceutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/klog/v2"
)

type usersGetter struct {
	aiInformer  informers.SharedInformerFactory
	k8sInformer k8sinformers.SharedInformerFactory
}

func New(aiInformer informers.SharedInformerFactory, k8sinformer k8sinformers.SharedInformerFactory) v1alpha2.Interface {
	return &usersGetter{aiInformer: aiInformer, k8sInformer: k8sinformer}
}

func (d *usersGetter) Get(_, name string) (runtime.Object, error) {
	return d.aiInformer.Iam().V1alpha2().Users().Lister().Get(name)
}

func (d *usersGetter) List(_ string, query *query.Query) (*api.ListResult, error) {

	var users []*iamv1alpha2.User
	var err error

	if namespace := query.Filters[iamv1alpha2.ScopeNamespace]; namespace != "" {
		role := query.Filters[iamv1alpha2.ResourcesSingularRole]
		users, err = d.listAllUsersInNamespace(string(namespace), string(role))
		delete(query.Filters, iamv1alpha2.ScopeNamespace)
		delete(query.Filters, iamv1alpha2.ResourcesSingularRole)
	} else if workspace := query.Filters[iamv1alpha2.ScopeWorkspace]; workspace != "" {
		workspaceRole := query.Filters[iamv1alpha2.ResourcesSingularWorkspaceRole]
		users, err = d.listAllUsersInWorkspace(string(workspace), string(workspaceRole))
		delete(query.Filters, iamv1alpha2.ScopeWorkspace)
		delete(query.Filters, iamv1alpha2.ResourcesSingularWorkspaceRole)
	} else if cluster := query.Filters[iamv1alpha2.ScopeCluster]; cluster == "true" {
		clusterRole := query.Filters[iamv1alpha2.ResourcesSingularClusterRole]
		users, err = d.listAllUsersInCluster(string(clusterRole))
		delete(query.Filters, iamv1alpha2.ScopeCluster)
		delete(query.Filters, iamv1alpha2.ResourcesSingularClusterRole)
	} else if globalRole := query.Filters[iamv1alpha2.ResourcesSingularGlobalRole]; globalRole != "" {
		users, err = d.listAllUsersByGlobalRole(string(globalRole))
		delete(query.Filters, iamv1alpha2.ResourcesSingularGlobalRole)
	} else {
		users, err = d.aiInformer.Iam().V1alpha2().Users().Lister().List(query.Selector())
	}

	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, user := range users {
		result = append(result, user)
	}

	return v1alpha2.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *usersGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftUser, ok := left.(*iamv1alpha2.User)
	if !ok {
		return false
	}

	rightUser, ok := right.(*iamv1alpha2.User)
	if !ok {
		return false
	}

	return v1alpha2.DefaultObjectMetaCompare(leftUser.ObjectMeta, rightUser.ObjectMeta, field)
}

func (d *usersGetter) filter(object runtime.Object, filter query.Filter) bool {
	user, ok := object.(*iamv1alpha2.User)

	if !ok {
		return false
	}

	switch filter.Field {
	case iamv1alpha2.FieldEmail:
		return user.Spec.Email == string(filter.Value)
	case iamv1alpha2.InGroup:
		return sliceutil.HasString(user.Spec.Groups, string(filter.Value))
	case iamv1alpha2.NotInGroup:
		return !sliceutil.HasString(user.Spec.Groups, string(filter.Value))
	default:
		return v1alpha2.DefaultObjectMetaFilter(user.ObjectMeta, filter)
	}
}

func (d *usersGetter) listAllUsersInWorkspace(workspace, role string) ([]*iamv1alpha2.User, error) {
	var users []*iamv1alpha2.User
	var err error
	workspaceRoleBindings, err := d.aiInformer.Iam().V1alpha2().
		WorkspaceRoleBindings().Lister().List(labels.SelectorFromValidatedSet(labels.Set{tenantv1alpha2.WorkspaceLabel: workspace}))

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, roleBinding := range workspaceRoleBindings {
		if role != "" && roleBinding.RoleRef.Name != role {
			continue
		}
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == iamv1alpha2.ResourceKindUser {

				if contains(users, subject.Name) {
					klog.Warningf("conflict role binding found: %s, username:%s", roleBinding.ObjectMeta.String(), subject.Name)
					continue
				}

				obj, err := d.Get("", subject.Name)

				if err != nil {
					if errors.IsNotFound(err) {
						klog.Warningf("orphan subject: %s", subject.String())
						continue
					}
					klog.Error(err)
					return nil, err
				}

				user := obj.(*iamv1alpha2.User)
				user = user.DeepCopy()
				if user.Annotations == nil {
					user.Annotations = make(map[string]string, 0)
				}
				user.Annotations[iamv1alpha2.WorkspaceRoleAnnotation] = roleBinding.RoleRef.Name
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func (d *usersGetter) listAllUsersInNamespace(namespace, role string) ([]*iamv1alpha2.User, error) {
	var users []*iamv1alpha2.User
	var err error

	roleBindings, err := d.k8sInformer.Rbac().V1().
		RoleBindings().Lister().RoleBindings(namespace).List(labels.Everything())

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, roleBinding := range roleBindings {
		if role != "" && roleBinding.RoleRef.Name != role {
			continue
		}
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == iamv1alpha2.ResourceKindUser {
				if contains(users, subject.Name) {
					klog.Warningf("conflict role binding found: %s, username:%s", roleBinding.ObjectMeta.String(), subject.Name)
					continue
				}

				obj, err := d.Get("", subject.Name)

				if err != nil {
					if errors.IsNotFound(err) {
						klog.Warningf("orphan subject: %s", subject.String())
						continue
					}
					klog.Error(err)
					return nil, err
				}

				user := obj.(*iamv1alpha2.User)
				user = user.DeepCopy()
				if user.Annotations == nil {
					user.Annotations = make(map[string]string, 0)
				}
				user.Annotations[iamv1alpha2.RoleAnnotation] = roleBinding.RoleRef.Name
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func (d *usersGetter) listAllUsersByGlobalRole(globalRole string) ([]*iamv1alpha2.User, error) {
	var users []*iamv1alpha2.User
	var err error

	globalRoleBindings, err := d.aiInformer.Iam().V1alpha2().
		GlobalRoleBindings().Lister().List(labels.Everything())

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, roleBinding := range globalRoleBindings {
		if roleBinding.RoleRef.Name != globalRole {
			continue
		}
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == iamv1alpha2.ResourceKindUser {

				if contains(users, subject.Name) {
					klog.Warningf("conflict role binding found: %s, username:%s", roleBinding.ObjectMeta.String(), subject.Name)
					continue
				}

				obj, err := d.Get("", subject.Name)

				if err != nil {
					if errors.IsNotFound(err) {
						klog.Warningf("orphan subject: %s", subject.String())
						continue
					}
					klog.Error(err)
					return nil, err
				}

				user := obj.(*iamv1alpha2.User)
				user = user.DeepCopy()
				if user.Annotations == nil {
					user.Annotations = make(map[string]string, 0)
				}
				user.Annotations[iamv1alpha2.GlobalRoleAnnotation] = roleBinding.RoleRef.Name
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func (d *usersGetter) listAllUsersInCluster(clusterRole string) ([]*iamv1alpha2.User, error) {
	var users []*iamv1alpha2.User
	var err error

	roleBindings, err := d.k8sInformer.Rbac().V1().ClusterRoleBindings().Lister().List(labels.Everything())

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, roleBinding := range roleBindings {
		if clusterRole != "" && roleBinding.RoleRef.Name != clusterRole {
			continue
		}
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == iamv1alpha2.ResourceKindUser {
				if contains(users, subject.Name) {
					klog.Warningf("conflict role binding found: %s, username:%s", roleBinding.ObjectMeta.String(), subject.Name)
					continue
				}

				obj, err := d.Get("", subject.Name)

				if err != nil {
					if errors.IsNotFound(err) {
						klog.Warningf("orphan subject: %s", subject.String())
						continue
					}
					klog.Error(err)
					return nil, err
				}

				user := obj.(*iamv1alpha2.User)
				user = user.DeepCopy()
				if user.Annotations == nil {
					user.Annotations = make(map[string]string, 0)
				}
				user.Annotations[iamv1alpha2.ClusterRoleAnnotation] = roleBinding.RoleRef.Name
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func contains(users []*iamv1alpha2.User, username string) bool {
	for _, user := range users {
		if user.Name == username {
			return true
		}
	}
	return false
}
