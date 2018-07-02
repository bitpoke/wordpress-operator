/*
Copyright 2018 Pressinfra SRL

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

package wordpress

import (
	apps_util "github.com/appscode/kutil/apps/v1"
	"github.com/golang/glog"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	wpapi "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	"github.com/presslabs/wordpress-operator/pkg/factory/wordpress"
)

const (
	deploymentName = "%s"
)

func (c *Controller) syncDeployment(wp *wpapi.Wordpress) error {
	glog.Infof("Syncing deployment for %s/%s", wp.ObjectMeta.Namespace, wp.ObjectMeta.Name)

	wpf := wordpress.Generator{
		WP:                  wp.WithDefaults(),
		DefaultRuntimeImage: c.RuntimeImage,
	}
	labels := wpf.Labels()
	labels["app.kubernetes.io/component"] = "web"
	labels["app.kubernetes.io/tier"] = "front"

	meta := c.objectMeta(wp, deploymentName)
	meta.Labels = labels

	_, _, err := apps_util.CreateOrPatchDeployment(c.KubeClient, meta, func(in *appsv1.Deployment) *appsv1.Deployment {
		in.ObjectMeta = c.ensureControllerReference(in.ObjectMeta, wp)

		in.Spec.Selector = metav1.SetAsLabelSelector(labels)
		in.Spec.Template = *wpf.PodTemplateSpec(&in.Spec.Template)
		in.Spec.Template.ObjectMeta.Labels = labels

		if wp.Spec.Replicas != nil {
			in.Spec.Replicas = wp.Spec.Replicas
		}

		return in
	})
	return err
}
