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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
)

const (
	EventReasonIngressFailed  EventReason = "IngressFailed"
	EventReasonIngressUpdated EventReason = "IngressUpdated"
)

type IngressSyncer struct {
	scheme   *runtime.Scheme
	wp       *wordpressv1alpha1.Wordpress
	key      types.NamespacedName
	existing *extv1beta1.Ingress
}

var _ Interface = &IngressSyncer{}

func NewIngressSyncer(wp *wordpressv1alpha1.Wordpress, r *runtime.Scheme) *IngressSyncer {
	return &IngressSyncer{
		scheme:   r,
		wp:       wp,
		existing: &extv1beta1.Ingress{},
		key: types.NamespacedName{
			Name:      wp.GetIngressName(),
			Namespace: wp.Namespace,
		},
	}
}

func (s *IngressSyncer) GetKey() types.NamespacedName                 { return s.key }
func (s *IngressSyncer) GetExistingObjectPlaceholder() runtime.Object { return s.existing }

func (s *IngressSyncer) T(in runtime.Object) (runtime.Object, error) {
	out := in.(*extv1beta1.Ingress)

	out.Name = s.key.Name
	out.Namespace = s.key.Namespace
	out.Labels = s.wp.WebPodLabels()
	if err := controllerutil.SetControllerReference(s.wp, out, s.scheme); err != nil {
		return nil, err
	}

	for k, v := range s.wp.Spec.IngressAnnotations {
		out.ObjectMeta.Annotations[k] = v
	}

	bk := extv1beta1.IngressBackend{
		ServiceName: s.wp.GetServiceName(),
		ServicePort: intstr.FromString("http"),
	}
	bkpaths := []extv1beta1.HTTPIngressPath{
		extv1beta1.HTTPIngressPath{
			Path:    "/",
			Backend: bk,
		},
	}

	rules := []extv1beta1.IngressRule{}
	for _, d := range s.wp.Spec.Domains {
		rules = append(rules, extv1beta1.IngressRule{
			Host: string(d),
			IngressRuleValue: extv1beta1.IngressRuleValue{
				HTTP: &extv1beta1.HTTPIngressRuleValue{
					Paths: bkpaths,
				},
			},
		})
	}
	out.Spec.Rules = rules

	if len(s.wp.Spec.TLSSecretRef) > 0 {
		tls := extv1beta1.IngressTLS{
			SecretName: string(s.wp.Spec.TLSSecretRef),
		}
		for _, d := range s.wp.Spec.Domains {
			tls.Hosts = append(tls.Hosts, string(d))
		}
		out.Spec.TLS = []extv1beta1.IngressTLS{tls}
	} else {
		out.Spec.TLS = []extv1beta1.IngressTLS{}
	}

	return out, nil
}

func (s *IngressSyncer) GetErrorEventReason(err error) EventReason {
	if err == nil {
		return EventReasonIngressUpdated
	} else {
		return EventReasonIngressFailed
	}
}
