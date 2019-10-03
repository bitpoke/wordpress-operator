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
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/presslabs/controller-util/syncer"

	"github.com/presslabs/wordpress-operator/pkg/cmd/options"
	"github.com/presslabs/wordpress-operator/pkg/internal/wordpress"
)

const ingressClassAnnotationKey = "kubernetes.io/ingress.class"

func upsertPath(rules []extv1beta1.IngressRule, domain, path string, bk extv1beta1.IngressBackend) []extv1beta1.IngressRule {
	hostIdx := -1
	for i := range rules {
		if rules[i].Host == domain {
			hostIdx = i
		}
	}
	if hostIdx == -1 {
		rules = append(rules, extv1beta1.IngressRule{Host: domain})
		hostIdx = len(rules) - 1
	}

	if rules[hostIdx].HTTP == nil {
		rules[hostIdx].HTTP = &extv1beta1.HTTPIngressRuleValue{}
	}

	idx := -1
	for i := range rules[hostIdx].HTTP.Paths {
		if rules[hostIdx].HTTP.Paths[i].Path == path {
			idx = i
		}
	}
	if idx == -1 {
		rules[hostIdx].HTTP.Paths = append(rules[hostIdx].HTTP.Paths,
			extv1beta1.HTTPIngressPath{Path: path})
		idx = len(rules[hostIdx].HTTP.Paths) - 1
	}

	rules[hostIdx].HTTP.Paths[idx].Backend = bk

	return rules
}

// NewIngressSyncer returns a new sync.Interface for reconciling web Ingress
func NewIngressSyncer(wp *wordpress.Wordpress, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	objLabels := wp.ComponentLabels(wordpress.WordpressIngress)

	obj := &extv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wp.ComponentName(wordpress.WordpressIngress),
			Namespace: wp.Namespace,
		},
	}

	bk := extv1beta1.IngressBackend{
		ServiceName: wp.ComponentName(wordpress.WordpressService),
		ServicePort: intstr.FromString("http"),
	}

	return syncer.NewObjectSyncer("Ingress", wp.Unwrap(), obj, c, scheme, func(existing runtime.Object) error {
		out := existing.(*extv1beta1.Ingress)
		out.Labels = labels.Merge(labels.Merge(out.Labels, objLabels), controllerLabels)

		if len(out.ObjectMeta.Annotations) == 0 && (len(wp.Spec.IngressAnnotations) > 0 || options.IngressClass != "") {
			out.ObjectMeta.Annotations = make(map[string]string)
		}
		if options.IngressClass != "" {
			out.ObjectMeta.Annotations[ingressClassAnnotationKey] = options.IngressClass
		}
		for k, v := range wp.Spec.IngressAnnotations {
			out.ObjectMeta.Annotations[k] = v
		}

		rules := []extv1beta1.IngressRule{}
		for _, route := range wp.Spec.Routes {
			path := route.Path
			if path == "" {
				path = "/"
			}
			rules = upsertPath(rules, route.Domain, path, bk)
		}

		out.Spec.Rules = rules

		if len(wp.Spec.TLSSecretRef) > 0 {
			tls := extv1beta1.IngressTLS{
				SecretName: string(wp.Spec.TLSSecretRef),
			}
			for _, route := range wp.Spec.Routes {
				tls.Hosts = append(tls.Hosts, route.Domain)
			}
			out.Spec.TLS = []extv1beta1.IngressTLS{tls}
		} else {
			out.Spec.TLS = nil
		}

		return nil
	})
}
