/*
Copyright 2025.

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

package v1beta

import (
	"github.com/cert-manager/issuer-lib/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:object:generate=true

// PCAClusterIssuer is the Schema for the pcaclusterissuers API.
// +kubebuilder:resource:path=pcaclusterissuers,scope=Cluster
type PCAClusterIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PCAIssuerSpec         `json:"spec,omitempty"`
	Status v1alpha1.IssuerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PCAClusterIssuerList contains a list of PCAClusterIssuer.
type PCAClusterIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PCAClusterIssuer `json:"items"`
}

// +kubebuilder:object:generate=true

// PCAIssuerSpec defines the desired state of PCAIssuer and PCAClusterIssuer
type PCAIssuerSpec struct {
	// ParentIdentifier is CA certificate identifier.
	ParentIdentifier         string     `json:"parent_identifier"`
	AccessKey                *SecretRef `json:"accessKey,omitempty"`
	AccessKeySecret          *SecretRef `json:"accessKeySecret,omitempty"`
	RAMRoleARN               string     `json:"ramRoleARN,omitempty"`
	RAMRoleSessionName       string     `json:"ramRoleSessionName,omitempty"`
	OIDCProviderARN          string     `json:"oidcProviderARN,omitempty"`
	OIDCTokenFilePath        string     `json:"oidcTokenFilePath,omitempty"`
	RoleSessionExpiration    string     `json:"roleSessionExpiration,omitempty"`
	RemoteRAMRoleARN         string     `json:"remoteRamRoleARN,omitempty"`
	RemoteRAMRoleSessionName string     `json:"remoteRamRoleSessionName,omitempty"`
}

type SecretRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:object:generate=true

// PCAIssuer is the Schema for the pcaissuers API.
// +kubebuilder:resource:path=pcaissuers,scope=Namespaced
type PCAIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PCAIssuerSpec         `json:"spec,omitempty"`
	Status v1alpha1.IssuerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PCAIssuerList contains a list of PCAIssuer.
type PCAIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PCAIssuer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PCAClusterIssuer{}, &PCAClusterIssuerList{})
	SchemeBuilder.Register(&PCAIssuer{}, &PCAIssuerList{})
}

func (p *PCAIssuer) GetStatus() *v1alpha1.IssuerStatus {
	return &(p.Status)
}

func (p *PCAIssuer) GetIssuerTypeIdentifier() string {
	// ACTION REQUIRED: Change this to a unique string that identifies your issuer
	return "pcaissuers.alibabacloud.com"
}

func (p *PCAClusterIssuer) GetStatus() *v1alpha1.IssuerStatus {
	return &(p.Status)
}

func (p *PCAClusterIssuer) GetIssuerTypeIdentifier() string {
	// ACTION REQUIRED: Change this to a unique string that identifies your issuer
	return "pcaclusterissuers.alibabacloud.com"
}

var _ v1alpha1.Issuer = &PCAClusterIssuer{}
var _ v1alpha1.Issuer = &PCAIssuer{}
