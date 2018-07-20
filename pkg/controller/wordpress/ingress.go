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
	ext_util "github.com/appscode/kutil/extensions/v1beta1"
	"github.com/golang/glog"
	extv1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"

	wpapi "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	"github.com/presslabs/wordpress-operator/pkg/factory/wordpress"
)

func (c *Controller) syncIngress(wp *wpapi.Wordpress) error {
	glog.Infof("Syncing ingress for %s/%s", wp.ObjectMeta.Namespace, wp.ObjectMeta.Name)

	wpf := wordpress.Generator{WP: wp}

	meta := c.objectMeta(wp, wpf.IngressName())

	_, _, err := ext_util.CreateOrPatchIngress(c.KubeClient, meta, func(in *extv1.Ingress) *extv1.Ingress {
		in.ObjectMeta = c.ensureControllerReference(in.ObjectMeta, wp)
		in.ObjectMeta.Annotations = wp.Spec.IngressAnnotations

		bk := extv1.IngressBackend{
			ServiceName: wpf.ServiceName(),
			ServicePort: intstr.FromString("http"),
		}
		bkpaths := []extv1.HTTPIngressPath{
			extv1.HTTPIngressPath{
				Path:    "/",
				Backend: bk,
			},
		}

		rules := []extv1.IngressRule{}
		for _, d := range wp.Spec.Domains {
			rules = append(rules, extv1.IngressRule{
				Host: string(d),
				IngressRuleValue: extv1.IngressRuleValue{
					HTTP: &extv1.HTTPIngressRuleValue{
						Paths: bkpaths,
					},
				},
			})
		}
		in.Spec.Rules = rules

		if len(wp.Spec.TLSSecretRef) > 0 {
			tls := extv1.IngressTLS{
				SecretName: string(wp.Spec.TLSSecretRef),
			}
			for _, d := range wp.Spec.Domains {
				tls.Hosts = append(tls.Hosts, string(d))
			}
			in.Spec.TLS = []extv1.IngressTLS{tls}
		} else {
			in.Spec.TLS = []extv1.IngressTLS{}
		}

		return in
	})

	return err
}
