package apis

import (
	tenantv1alpha2 "aiscope/pkg/apis/tenant/v1alpha2"
)

func init() {
	AddToSchemes = append(AddToSchemes, tenantv1alpha2.SchemeBuilder.AddToScheme)
}
