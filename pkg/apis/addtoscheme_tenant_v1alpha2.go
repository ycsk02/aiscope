package apis

import (
	tenantv1alpha2 "aiscope/pkg/api/tenant/v1alpha2"
)

func init() {
	AddToSchemes = append(AddToSchemes, tenantv1alpha2.SchemeBuilder.AddToScheme)
}
