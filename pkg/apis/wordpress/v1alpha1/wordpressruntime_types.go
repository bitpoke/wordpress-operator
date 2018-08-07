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

// WordpressRuntimeSpec defines the desired state of WordpressRuntime
type WordpressRuntimeSpec struct {
	// DefaultImage is the image for the placeholder image. This image is used
	// in containers that specify "defaultImage" as their image
	// +kubebuilder:validation:MinLength=1
	DefaultImage string `json:"defaultImage"`
	// DefaultPullPolicyImage is the pull policy which gets set for the
	// defaultImage
	// +kubebuilder:validation:Enum=Always,IfNotPresent,Never
	// +optional
	DefaultImagePullPolicy corev1.PullPolicy `json:"defaultImagePullPolicy,omitempty"`
	// WebrootVolumeSpec defines the volume for storing the wordpress
	// installation.
	// +optional
	WebrootVolumeSpec *WordpressVolumeSpec `json:"webrootVolumeSpec,omitempty"`
	// MediaVolumeSpec if specified, defines a separate volume for storing
	// media files.
	// +optional
	MediaVolumeSpec *WordpressVolumeSpec `json:"mediaVolumeSpec,omitempty"`
	// WebPodTemplate is the pod template for the WordPress web frontend.
	//
	//
	// *The globally defined volume mounts* are injected into all containers
	//
	// *The globally defined env* is injected into all containers
	WebPodTemplate *corev1.PodTemplateSpec `json:"webPodTemplate"`
	// CLIPodTemplate is the pod template for running wp-cli commands (eg.
	// wp-cron, wp database upgrades, etc.)
	//
	// *The globally defined volume mounts* are injected into all containers
	//
	// *The globally defined env* is injected into all containers
	//
	// The pod restart policy is set `Never`, regardless of the spec
	//
	CLIPodTemplate *corev1.PodTemplateSpec `json:"cliPodTemplate"`
	// If specified apply these annotations to the Ingress resource created for
	// this Wordpress Site.
	// +optional
	IngressAnnotations map[string]string `json:"ingressAnnotations,omitempty"`
	// ServiceSpec is the specification for the service created for this
	// WordPress Site
	// By default, a ClusterIP service which exposes http port of web pods
	// +optional
	ServiceSpec *corev1.ServiceSpec `json:"serviceSpec,omitempty"`
}

// WordpressRuntimeStatus defines the observed state of WordpressRuntime
type WordpressRuntimeStatus struct{}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// WordpressRuntime is the Schema for the wordpressruntimes API
// +k8s:openapi-gen=true
type WordpressRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WordpressRuntimeSpec   `json:"spec,omitempty"`
	Status WordpressRuntimeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// WordpressRuntimeList contains a list of WordpressRuntime
type WordpressRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WordpressRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WordpressRuntime{}, &WordpressRuntimeList{})
}
