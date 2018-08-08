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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
)

const (
	// EventReasonServiceFailed is the event reason for a failed Service reconcile
	EventReasonServiceFailed EventReason = "ServiceFailed"
	// EventReasonServiceUpdated is the event reason for a successful Service reconcile
	EventReasonServiceUpdated EventReason = "ServiceUpdated"
)

type serviceSyncer struct {
	scheme   *runtime.Scheme
	wp       *wordpressv1alpha1.Wordpress
	rt       *wordpressv1alpha1.WordpressRuntime
	key      types.NamespacedName
	existing *corev1.Service
}

// NewServiceSyncer returns a new sync.Interface for reconciling web Service
func NewServiceSyncer(wp *wordpressv1alpha1.Wordpress, rt *wordpressv1alpha1.WordpressRuntime, r *runtime.Scheme) Interface {
	return &serviceSyncer{
		scheme:   r,
		wp:       wp,
		rt:       rt,
		existing: &corev1.Service{},
		key: types.NamespacedName{
			Name:      wp.GetServiceName(),
			Namespace: wp.Namespace,
		},
	}
}

func (s *serviceSyncer) GetKey() types.NamespacedName                 { return s.key }
func (s *serviceSyncer) GetExistingObjectPlaceholder() runtime.Object { return s.existing }

func (s *serviceSyncer) T(in runtime.Object) (runtime.Object, error) {
	out := in.(*corev1.Service)

	out.Name = s.key.Name
	out.Namespace = s.key.Namespace
	out.Labels = s.wp.WebPodLabels()
	if err := controllerutil.SetControllerReference(s.wp, out, s.scheme); err != nil {
		return nil, err
	}

	inspec := out.Spec.DeepCopy()

	out.Spec = *s.rt.Spec.ServiceSpec

	// Spec.ClusterIP of an service is immutable
	if len(inspec.ClusterIP) > 0 {
		out.Spec.ClusterIP = inspec.ClusterIP
	}

	out.Spec.Selector = s.wp.WebPodLabels()

	return out, nil
}

func (s *serviceSyncer) GetErrorEventReason(err error) EventReason {
	if err != nil {
		return EventReasonServiceFailed
	}
	return EventReasonServiceUpdated
}
