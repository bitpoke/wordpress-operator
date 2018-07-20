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
	core_util "github.com/appscode/kutil/core/v1"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	wpapi "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	"github.com/presslabs/wordpress-operator/pkg/factory/wordpress"
)

const ()

func (c *Controller) syncService(wp *wpapi.Wordpress) error {
	glog.Infof("Syncing service for %s/%s", wp.ObjectMeta.Namespace, wp.ObjectMeta.Name)

	wpf := wordpress.Generator{WP: wp}
	l := wpf.WebPodLabels()

	meta := c.objectMeta(wp, wpf.ServiceName())
	meta.Labels = l

	_, _, err := core_util.CreateOrPatchService(c.KubeClient, meta, func(in *corev1.Service) *corev1.Service {
		in.ObjectMeta = c.ensureControllerReference(in.ObjectMeta, wp)

		if len(in.Spec.Type) == 0 {
			in.Spec.Type = corev1.ServiceTypeClusterIP
		}

		if wp.Spec.ServiceSpec != nil {
			in.Spec.LoadBalancerIP = wp.Spec.ServiceSpec.LoadBalancerIP
			in.Spec.ExternalIPs = wp.Spec.ServiceSpec.ExternalIPs
			in.Spec.ExternalName = wp.Spec.ServiceSpec.ExternalName

			in.Spec.LoadBalancerSourceRanges = wp.Spec.ServiceSpec.LoadBalancerSourceRanges

			if wp.Spec.ServiceSpec.HealthCheckNodePort > 0 {
				in.Spec.HealthCheckNodePort = wp.Spec.ServiceSpec.HealthCheckNodePort
			}
			if wp.Spec.ServiceSpec.SessionAffinityConfig != nil {
				in.Spec.SessionAffinityConfig = wp.Spec.ServiceSpec.SessionAffinityConfig.DeepCopy()
			}

			if len(wp.Spec.ServiceSpec.ExternalTrafficPolicy) > 0 {
				in.Spec.ExternalTrafficPolicy = wp.Spec.ServiceSpec.ExternalTrafficPolicy
			}

			if len(wp.Spec.ServiceSpec.Type) > 0 {
				in.Spec.Type = wp.Spec.ServiceSpec.Type
			}
			if len(wp.Spec.ServiceSpec.SessionAffinity) > 0 {
				in.Spec.SessionAffinity = wp.Spec.ServiceSpec.SessionAffinity
			}
			if len(wp.Spec.ServiceSpec.ClusterIP) > 0 {
				in.Spec.ClusterIP = wp.Spec.ServiceSpec.ClusterIP
			}

			in.Spec.PublishNotReadyAddresses = false
		} else {
			in.Spec.Type = corev1.ServiceTypeClusterIP
		}

		in.Spec.Selector = l

		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []corev1.ServicePort{
			corev1.ServicePort{
				Name:       "http",
				Port:       80,
				TargetPort: intstr.FromString("http"),
			},
		})

		// make sure we remove the NodePort if the service is of type ClusterIP
		if in.Spec.Type == corev1.ServiceTypeClusterIP {
			for i, p := range in.Spec.Ports {
				if p.Name == "http" {
					in.Spec.Ports[i].NodePort = 0
				}
			}
		}

		return in
	})

	return err
}
