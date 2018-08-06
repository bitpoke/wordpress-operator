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
	// Number of desired web pods. This is a pointer to distinguish between
	// explicit zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Domains for this this site answers. The first item is set as the "main
	// domain" (WP_HOME and WP_SITEURL constants).
	Domains []Domain `json:"domains"`
	// TLSSecretRef a secret containing the TLS certificates for this site.
	// +optional
	TLSSecretRef SecretRef `json:"tlsSecretRef,omitempty"`
	// WebrootVolumeSpec defines the volume for storing the wordpress
	// installation.
	// +optional
	WebrootVolumeSpec WordpressVolumeSpec `json:"webrootVolumeSpec,omitempty"`
	// MediaVolumeSpec if specified, defines a separate volume for storing
	// media files.
	// +optional
	MediaVolumeSpec *WordpressVolumeSpec `json:"mediaVolumeSpec,omitempty"`
	// VolumeMountsSpec defines the mount structure for mounting volumes into
	// pods. Each container in WebPodTemplate and CLIPodTemplate will get this
	// structure mounted.
	// If undefined, WebrootVolume gets mounted into /var/www/html/ and
	// if defined, the MediaVolume gets mounted into /var/www/html/wp-content/uploads
	// +optional
	VolumeMountsSpec []corev1.VolumeMount `json:"volumeMountsSpec,omitempty"`
	// Env that gets injected into every container of WebPodTemplate and
	// CLIPodTemplate
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	// EnvFrom gets injected into every container of WebPodTemplate and
	// CLIPodTemplate
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
	// WebPodTemplate is the pod template for the WordPress web frontend.
	//
	// *The globally defined volume mounts* are injected into all containers
	//
	// *The globally defined env* is injected into all containers
	WebPodTemplate *corev1.PodTemplateSpec `json:"webPodTemplate,omitempty"`
	// CLIPodTemplate is the pod template for running wp-cli commands (eg.
	// wp-cron, wp database upgrades, etc.)
	//
	// *The globally defined volume mounts* are injected into all containers
	//
	// *The globally defined env* is injected into all containers
	//
	// The pod restart policy is set `Never`, regardless of the spec
	//
	CLIPodTemplate *corev1.PodTemplateSpec `json:"cliPodTemplate,omitempty"`
	// If specified apply these annotations to the Ingress resource created for
	// this Wordpress Site.
	// +optional
	IngressAnnotations map[string]string `json:"ingressAnnotations,omitempty"`
	// ServiceSpec is the specification for the service created for this
	// WordPress Site
	// +optional
	ServiceSpec *corev1.ServiceSpec `json:"serviceSpec,omitempty"`
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
