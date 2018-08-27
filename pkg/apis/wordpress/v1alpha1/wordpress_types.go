/*
Copyright 2018 Pressinfra SRL.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecretRef represents a reference to a Secret
type SecretRef string

// Domain represents a valid domain name
type Domain string

// WordpressSpec defines the desired state of Wordpress
type WordpressSpec struct {
	// WordpressRuntime to use
	// +kubebuilder:validation:MinLength=1
	Runtime string `json:"runtime"`
	// Number of desired web pods. This is a pointer to distinguish between
	// explicit zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Image overrides WordpressRuntime spec.defaultImage
	// +optional
	Image string `json:"image,omitempty"`
	// ImagePullPolicy overrides WordpressRuntime spec.imagePullPolicy
	// +kubebuilder:validation:Enum=Always,IfNotPresent,Never
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// ImagePullSecrets defines additional secrets to use when pulling images
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run this
	// site's pods
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Domains for which this this site answers.
	// The first item is set as the "main domain" (eg. WP_HOME and WP_SITEURL constants).
	// +kubebuilder:validation:MinItems=1
	Domains []Domain `json:"domains"`
	// TLSSecretRef a secret containing the TLS certificates for this site.
	// +optional
	TLSSecretRef SecretRef `json:"tlsSecretRef,omitempty"`
	// WebrootVolumeSpec overrides WordpressRuntime spec.webrootVolumeSpec
	// This field is immutable.
	// +optional
	WebrootVolumeSpec *WordpressVolumeSpec `json:"webrootVolumeSpec,omitempty"`
	// MediaVolumeSpec overrides WordpressRuntime spec.mediaVolumeSpec
	// This field is immutable.
	// +optional
	MediaVolumeSpec *WordpressVolumeSpec `json:"mediaVolumeSpec,omitempty"`
	// Volumes defines additional volumes to get injected into web and cli pods
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
	// VolumeMountsSpec defines additional mounts which get injected into web
	// and cli pods.
	// +optional
	VolumeMountsSpec []corev1.VolumeMount `json:"volumeMountsSpec,omitempty"`
	// Env defines additional environment variables which get injected into web
	// and cli pods
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	// EnvFrom defines additional envFrom's which get injected into web
	// and cli pods
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
	// IngressAnnotations for this Wordpress site
	// +optional
	IngressAnnotations map[string]string `json:"ingressAnnotations,omitempty"`
	// Labels to apply to generated resources
	Labels map[string]string `json:"labels,omitempty"`
}

// WordpressVolumeSpec is the desired spec of a wordpress volume
type WordpressVolumeSpec struct {
	// EmptyDir to use if no PersistentVolumeClaim or HostPath is specified
	// +optional
	EmptyDir *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
	// HostPath to use instead of a PersistentVolumeClaim.
	// +optional
	HostPath *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`
	// PersistentVolumeClaim to use. It has the highest level of precedence,
	// followed by HostPath and EmptyDir
	// +optional
	PersistentVolumeClaim *corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
}

// WordpressStatus defines the observed state of Wordpress
type WordpressStatus struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Wordpress is the Schema for the wordpresses API
// +k8s:openapi-gen=true
type Wordpress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WordpressSpec   `json:"spec,omitempty"`
	Status WordpressStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WordpressList contains a list of Wordpress
type WordpressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Wordpress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Wordpress{}, &WordpressList{})
}
