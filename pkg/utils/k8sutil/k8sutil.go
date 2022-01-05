package k8sutil

import (
	tenantv1alpha2 "aiscope/pkg/apis/tenant/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsControlledBy returns whether the ownerReferences contains the specified resource kind
func IsControlledBy(ownerReferences []metav1.OwnerReference, kind string, name string) bool {
	for _, owner := range ownerReferences {
		if owner.Kind == kind && (name == "" || owner.Name == name) {
			return true
		}
	}
	return false
}

// RemoveWorkspaceOwnerReference remove workspace kind owner reference
func RemoveWorkspaceOwnerReference(ownerReferences []metav1.OwnerReference) []metav1.OwnerReference {
	tmp := make([]metav1.OwnerReference, 0)
	for _, owner := range ownerReferences {
		if owner.Kind != tenantv1alpha2.ResourceKindWorkspace {
			tmp = append(tmp, owner)
		}
	}
	return tmp
}

// GetWorkspaceOwnerName return workspace kind owner name
func GetWorkspaceOwnerName(ownerReferences []metav1.OwnerReference) string {
	for _, owner := range ownerReferences {
		if owner.Kind == tenantv1alpha2.ResourceKindWorkspace {
			return owner.Name
		}
	}
	return ""
}
