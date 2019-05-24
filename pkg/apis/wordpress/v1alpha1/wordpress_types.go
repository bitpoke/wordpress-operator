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

// SecretRef represents a reference to a Secret
type SecretRef string

// Domain represents a valid domain name
type Domain string

// WordpressConditionType defines condition types of a backup resources
type WordpressConditionType string

// WordpressCondition defines condition struct for backup resource
type WordpressCondition struct {
	// Type of Wordpress condition.
	Type WordpressConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// The reason for the condition's last transition.
	Reason string `json:"reason"`
	// A human readable message indicating details about the transition.
	Message string `json:"message"`
}

const (
	// WordpressConditionProvisioning states if a given Wordpress is still provisioning.
	WordpressConditionProvisioning WordpressConditionType = "Provisioning"
	// WordpressConditionError states if a given Wordpress has encountered an error.
	WordpressConditionError WordpressConditionType = "Error"
	// WordpressConditionRunning states if a given Wordpress is running.
	WordpressConditionRunning WordpressConditionType = "Running"

	// ProvisionInProgress is the reason associated to the Provisioning condition for when
	// the Wordpress provisioning is ongoing
	ProvisionInProgress string = "ProvisionInProgress"
	// ProvisionFailed is the reason associated to the Provisioning condition for when
	// the Wordpress provisioning has failed
	ProvisionFailed string = "ProvisionFailed"
	// ProvisionSuccessful is the reason associated to the Provisioning condition for when
	// the Wordpress provisioning has been successful
	ProvisionSuccessful string = "ProvisionSuccessful"
)

// WordpressSpec defines the desired state of Wordpress
type WordpressSpec struct {
	// Number of desired web pods. This is a pointer to distinguish between
	// explicit zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Domains for which this this site answers.
	// The first item is set as the "main domain" (eg. WP_HOME and WP_SITEURL constants).
	// +kubebuilder:validation:MinItems=1
	Domains []Domain `json:"domains"`
	// WordPress runtime image to use. Defaults to quay.io/presslabs/wordpress-runtime
	// +optional
	Image string `json:"image,omitempty"`
	// Image tag to use. Defaults to latest
	// +optional
	Tag string `json:"tag,omitempty"`
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
	// TLSSecretRef a secret containing the TLS certificates for this site.
	// +optional
	TLSSecretRef SecretRef `json:"tlsSecretRef,omitempty"`
	// CodeVolumeSpec specifies how the site's code gets mounted into the
	// container. If not specified, a code volume won't get mounted at all.
	// +optional
	CodeVolumeSpec *CodeVolumeSpec `json:"code,omitempty"`
	// MediaVolumeSpec specifies how media files get mounted into the runtime
	// container. If not specified, a media volume won't be mounted at all.
	// +optional
	MediaVolumeSpec *MediaVolumeSpec `json:"media,omitempty"`
	// Volumes defines additional volumes to get injected into web and cli pods
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
	// VolumeMountsSpec defines additional mounts which get injected into web
	// and cli pods.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// Env defines environment variables which get passed into web and cli pods
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	// EnvFrom defines envFrom's which get passed into web and cli containers
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
	// If specified, the resources required by wordpress container.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// If specified, Pod node selector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// If specified, indicates the pod's priority class
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// IngressAnnotations for this Wordpress site
	// +optional
	IngressAnnotations map[string]string `json:"ingressAnnotations,omitempty"`
}

// GitVolumeSource is the desired spec for git code source
type GitVolumeSource struct {
	// Repository is the git repository for the code
	Repository string `json:"repository"`
	// GitRef to clone (can be a branch name, but it should point to a tag or a
	// commit hash)
	// +optional
	GitRef string `json:"reference,omitempty"`
	// Env defines env variables  which get passed to the git clone container
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	// EnvFrom defines envFrom which get passed to the git clone container
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
	// EmptyDir volume to use for git cloning.
	// +optional
	EmptyDir *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
}

// S3VolumeSource is the desired spec for accessing media files over S3
// compatible object store
type S3VolumeSource struct {
	// Bucket for storing media files
	// +kubebuilder:validation:MinLength=1
	Bucket string `json:"bucket"`
	// PathPrefix is the prefix for media files in bucket
	PathPrefix string `json:"prefix,omitempty"`
	// Env variables for accessing S3 bucket. Taken into account are:
	// ACCESS_KEY, SECRET_ACCESS_KEY
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// GCSVolumeSource is the desired spec for accessing media files using google
// cloud storage object store
type GCSVolumeSource struct {
	// Bucket for storing media files
	// +kubebuilder:validation:MinLength=1
	Bucket string `json:"bucket"`
	// PathPrefix is the prefix for media files in bucket
	PathPrefix string `json:"prefix,omitempty"`
	// Env variables for accessing gcs bucket. Taken into account are:
	// GOOGLE_APPLICATION_CREDENTIALS_JSON
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// CodeVolumeSpec is the desired spec for mounting code into the wordpress
// runtime container
type CodeVolumeSpec struct {
	// ReadOnly specifies if the volume should be mounted read-only inside the
	// wordpress runtime container
	ReadOnly bool
	// MountPath spechfies where should the code volume be mounted.
	// Defaults to /var/www/site/web/wp-content
	// +optional
	MountPath string `json:"mountPath,omitempty"`
	// ContentSubPath specifies where within the code volumes, the wp-content
	// folder resides.
	// Defaults to wp-content/
	// +optional
	ContentSubPath string `json:"contentSubPath,omitempty"`
	// GitDir specifies the git repo to use for code cloning. It has the highest
	// level of precedence over EmptyDir, HostPath and PersistentVolumeClaim
	// +optional
	GitDir *GitVolumeSource `json:"git,omitempty"`
	// PersistentVolumeClaim to use if no GitDir is specified
	// +optional
	PersistentVolumeClaim *corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
	// HostPath to use if no PersistentVolumeClaim is specified
	// +optional
	HostPath *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`
	// EmptyDir to use if no HostPath is specified
	// +optional
	EmptyDir *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
}

// MediaVolumeSpec is the desired spec for mounting code into the wordpress
// runtime container
type MediaVolumeSpec struct {
	// ReadOnly specifies if the volume should be mounted read-only inside the
	// wordpress runtime container
	ReadOnly bool
	// S3VolumeSource specifies the S3 object storage configuration for media
	// files. It has the highest level of precedence over EmptyDir, HostPath
	// and PersistentVolumeClaim
	// +optional
	S3VolumeSource *S3VolumeSource `json:"s3,omitempty"`
	// GCSVolumeSource specifies the google cloud storage object storage
	// configuration for media files. It has the highest level of precedence
	// over EmptyDir, HostPath and PersistentVolumeClaim
	// +optional
	GCSVolumeSource *GCSVolumeSource `json:"gcs,omitempty"`
	// PersistentVolumeClaim to use if no S3VolumeSource or GCSVolumeSource are
	// specified
	// +optional
	PersistentVolumeClaim *corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
	// HostPath to use if no PersistentVolumeClaim is specified
	// +optional
	HostPath *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`
	// EmptyDir to use if no HostPath is specified
	// +optional
	EmptyDir *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
}

// WordpressStatus defines the observed state of Wordpress
type WordpressStatus struct {
	// Conditions represents the Wordpress resource conditions list.
	// +optional
	Conditions []WordpressCondition `json:"conditions,omitempty"`
	// Total number of non-terminated pods targeted by web deployment
	// This is copied over from the deployment object
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Wordpress is the Schema for the wordpresses API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
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
