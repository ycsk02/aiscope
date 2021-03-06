/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LoginRecordSpec defines the desired state of LoginRecord
type LoginRecordSpec struct {
	// Which authentication method used, BasicAuth/OAuth
	Type LoginType `json:"type"`
	// Provider of authentication, Ldap/Github etc.
	Provider string `json:"provider"`
	// Source IP of client
	SourceIP string `json:"sourceIP"`
	// User agent of login attempt
	UserAgent string `json:"userAgent,omitempty"`
	// Successful login attempt or not
	Success bool `json:"success"`
	// States failed login attempt reason
	Reason string `json:"reason"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.provider"
// +kubebuilder:printcolumn:name="From",type="string",JSONPath=".spec.sourceIP"
// +kubebuilder:printcolumn:name="Success",type="string",JSONPath=".spec.success"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".spec.reason"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories="iam",scope="Cluster"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoginRecord is the Schema for the loginrecords API
type LoginRecord struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoginRecordSpec   `json:"spec,omitempty"`
}

type LoginType string

const (
	BasicAuth LoginType = "Basic"
	OAuth     LoginType = "OAuth"
	Token     LoginType = "Token"
)

//+kubebuilder:object:root=true

// LoginRecordList contains a list of LoginRecord
type LoginRecordList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoginRecord `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LoginRecord{}, &LoginRecordList{})
}
