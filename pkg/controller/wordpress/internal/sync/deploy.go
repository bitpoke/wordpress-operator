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
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/appscode/mergo"

	"github.com/presslabs/controller-util/mergo/transformers"
	"github.com/presslabs/controller-util/syncer"

	"github.com/presslabs/wordpress-operator/pkg/internal/wordpress"
)

var errImmutableDeploymentSelector = errors.New("deployment selector is immutable")

// NewDeploymentSyncer returns a new sync.Interface for reconciling web Deployment.
func NewDeploymentSyncer(wp *wordpress.Wordpress, secret *corev1.Secret, c client.Client) syncer.Interface {
	objLabels := wp.ComponentLabels(wordpress.WordpressDeployment)

	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wp.ComponentName(wordpress.WordpressDeployment),
			Namespace: wp.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Deployment", wp.Unwrap(), obj, c, func() error {
		obj.Labels = labels.Merge(labels.Merge(obj.Labels, objLabels), controllerLabels)

		template := wp.WebPodTemplateSpec()

		if len(template.Annotations) == 0 {
			template.Annotations = make(map[string]string)
		}
		template.Annotations["wordpress.presslabs.org/secretVersion"] = secret.ResourceVersion

		obj.Spec.Template.ObjectMeta = template.ObjectMeta

		selector := metav1.SetAsLabelSelector(wp.WebPodLabels())
		if !reflect.DeepEqual(selector, obj.Spec.Selector) {
			if obj.ObjectMeta.CreationTimestamp.IsZero() {
				obj.Spec.Selector = selector
			} else {
				return errImmutableDeploymentSelector
			}
		}

		err := mergo.Merge(&obj.Spec.Template.Spec, template.Spec, mergo.WithTransformers(transformers.PodSpec))
		if err != nil {
			return err
		}

		obj.Spec.Template.Spec.NodeSelector = wp.Spec.NodeSelector
		obj.Spec.Template.Spec.Tolerations = wp.Spec.Tolerations

		if wp.Spec.Replicas != nil {
			obj.Spec.Replicas = wp.Spec.Replicas
		}

		if wp.Spec.DeploymentStrategy != nil {
			obj.Spec.Strategy = *wp.Spec.DeploymentStrategy
		}

		return nil
	})
}
