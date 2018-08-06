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
package sync

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
)

const (
	EventReasonDeploymentFailed  EventReason = "DeploymentFailed"
	EventReasonDeploymentUpdated EventReason = "DeploymentUpdated"
)

type DeploymentSyncer struct {
	scheme   *runtime.Scheme
	wp       *wordpressv1alpha1.Wordpress
	key      types.NamespacedName
	existing *appsv1.Deployment
}

var _ Interface = &DeploymentSyncer{}

func NewDeploymentSyncer(wp *wordpressv1alpha1.Wordpress, r *runtime.Scheme) *DeploymentSyncer {
	return &DeploymentSyncer{
		scheme:   r,
		wp:       wp,
		existing: &appsv1.Deployment{},
		key: types.NamespacedName{
			Name:      wp.GetDeploymentName(),
			Namespace: wp.Namespace,
		},
	}
}

func (s *DeploymentSyncer) GetKey() types.NamespacedName                 { return s.key }
func (s *DeploymentSyncer) GetExistingObjectPlaceholder() runtime.Object { return s.existing }

func (s *DeploymentSyncer) T(in runtime.Object) (runtime.Object, error) {
	out := in.(*appsv1.Deployment)

	out.Name = s.key.Name
	out.Namespace = s.key.Namespace
	out.Labels = s.wp.WebPodLabels()
	if err := controllerutil.SetControllerReference(s.wp, out, s.scheme); err != nil {
		return nil, err
	}

	out.Spec.Selector = metav1.SetAsLabelSelector(s.wp.WebPodLabels())
	out.Spec.Template = *s.wp.WebPodTemplateSpec()
	if s.wp.Spec.Replicas != nil {
		out.Spec.Replicas = s.wp.Spec.Replicas
	}

	return out, nil
}

func (s *DeploymentSyncer) GetErrorEventReason(err error) EventReason {
	if err == nil {
		return EventReasonDeploymentUpdated
	} else {
		return EventReasonDeploymentFailed
	}
}
