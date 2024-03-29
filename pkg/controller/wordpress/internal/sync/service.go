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
	"errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/presslabs/controller-util/syncer"

	"github.com/bitpoke/wordpress-operator/pkg/internal/wordpress"
)

var errImmutableServiceSelector = errors.New("service selector is immutable")

// NewServiceSyncer returns a new sync.Interface for reconciling web Service.
func NewServiceSyncer(wp *wordpress.Wordpress, c client.Client) syncer.Interface {
	objLabels := wp.ComponentLabels(wordpress.WordpressDeployment)

	obj := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wp.Name,
			Namespace: wp.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Service", wp.Unwrap(), obj, c, func() error {
		obj.Labels = labels.Merge(labels.Merge(obj.Labels, objLabels), controllerLabels)

		selector := wp.WebPodLabels()
		if !labels.Equals(selector, obj.Spec.Selector) {
			if obj.ObjectMeta.CreationTimestamp.IsZero() {
				obj.Spec.Selector = selector
			} else {
				return errImmutableServiceSelector
			}
		}

		if len(obj.Spec.Ports) != 2 {
			obj.Spec.Ports = make([]corev1.ServicePort, 2)
		}

		obj.Spec.Ports[0].Name = "http"
		obj.Spec.Ports[0].Port = int32(80)
		obj.Spec.Ports[0].TargetPort = intstr.FromInt(wordpress.InternalHTTPPort)

		obj.Spec.Ports[1].Name = "prometheus"
		obj.Spec.Ports[1].Port = int32(wordpress.MetricsExporterPort)
		obj.Spec.Ports[1].TargetPort = intstr.FromInt(wordpress.MetricsExporterPort)

		return nil
	})
}
