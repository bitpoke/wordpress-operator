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
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/presslabs/controller-util/syncer"

	"github.com/bitpoke/wordpress-operator/pkg/cmd/options"
	"github.com/bitpoke/wordpress-operator/pkg/internal/wordpress"
)

const ingressClassAnnotationKey = "kubernetes.io/ingress.class"

func upsertPath(rules []netv1.IngressRule, domain, path string, bk netv1.IngressBackend) []netv1.IngressRule {
	var rule *netv1.IngressRule

	for i := range rules {
		if rules[i].Host == domain {
			rule = &rules[i]

			break
		}
	}

	if rule == nil {
		rules = append(rules, netv1.IngressRule{Host: domain})
		rule = &rules[len(rules)-1]
	}

	if rule.HTTP == nil {
		rule.HTTP = &netv1.HTTPIngressRuleValue{}
	}

	var httpPath *netv1.HTTPIngressPath

	for i := range rule.HTTP.Paths {
		if rule.HTTP.Paths[i].Path == path {
			httpPath = &rule.HTTP.Paths[i]

			break
		}
	}

	if httpPath == nil {
		rule.HTTP.Paths = append(rule.HTTP.Paths, netv1.HTTPIngressPath{Path: path})
		httpPath = &rule.HTTP.Paths[len(rule.HTTP.Paths)-1]
	}

	pathType := netv1.PathTypePrefix
	httpPath.PathType = &pathType
	httpPath.Backend = bk

	return rules
}

// NewIngressSyncer returns a new sync.Interface for reconciling web Ingress.
func NewIngressSyncer(wp *wordpress.Wordpress, c client.Client) syncer.Interface {
	objLabels := wp.ComponentLabels(wordpress.WordpressIngress)

	obj := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wp.ComponentName(wordpress.WordpressIngress),
			Namespace: wp.Namespace,
		},
	}

	bk := netv1.IngressBackend{
		Service: &netv1.IngressServiceBackend{
			Name: wp.ComponentName(wordpress.WordpressService),
			Port: netv1.ServiceBackendPort{Name: "http"},
		},
	}

	return syncer.NewObjectSyncer("Ingress", wp.Unwrap(), obj, c, func() error {
		obj.Labels = labels.Merge(labels.Merge(obj.Labels, objLabels), controllerLabels)

		if len(obj.ObjectMeta.Annotations) == 0 {
			obj.ObjectMeta.Annotations = make(map[string]string)
		}

		for k, v := range wp.Spec.IngressAnnotations {
			obj.ObjectMeta.Annotations[k] = v
		}
		delete(obj.ObjectMeta.Annotations, ingressClassAnnotationKey)

		if options.IngressClass != "" {
			obj.Spec.IngressClassName = &options.IngressClass
		} else {
			obj.Spec.IngressClassName = nil
		}

		rules := []netv1.IngressRule{}
		for _, route := range wp.Spec.Routes {
			path := route.Path
			if path == "" {
				path = "/"
			}
			rules = upsertPath(rules, route.Domain, path, bk)
		}

		obj.Spec.Rules = rules

		if len(wp.Spec.TLSSecretRef) > 0 {
			tls := netv1.IngressTLS{
				SecretName: string(wp.Spec.TLSSecretRef),
			}
			for _, route := range wp.Spec.Routes {
				tls.Hosts = append(tls.Hosts, route.Domain)
			}
			obj.Spec.TLS = []netv1.IngressTLS{tls}
		} else {
			obj.Spec.TLS = nil
		}

		return nil
	})
}
