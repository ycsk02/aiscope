package apis

import (
	experimentv1alpha2 "aiscope/pkg/apis/experiment/v1alpha2"
)

func init() {
	AddToSchemes = append(AddToSchemes, experimentv1alpha2.SchemeBuilder.AddToScheme)
}
