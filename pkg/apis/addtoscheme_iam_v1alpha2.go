package apis

import (
	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
)

func init() {
	AddToSchemes = append(AddToSchemes, iamv1alpha2.SchemeBuilder.AddToScheme)
}
