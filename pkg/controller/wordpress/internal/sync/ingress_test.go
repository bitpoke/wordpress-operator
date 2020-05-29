/*
Copyright 2019 Pressinfra SRL.

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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	netv1beta1 "k8s.io/api/networking/v1beta1"
)

var _ = Describe("The upsertPath function", func() {
	var (
		rules []netv1beta1.IngressRule
		bk    netv1beta1.IngressBackend
	)

	BeforeEach(func() {
		rules = []netv1beta1.IngressRule{
			{
				Host: "presslabs.com",
				IngressRuleValue: netv1beta1.IngressRuleValue{
					HTTP: &netv1beta1.HTTPIngressRuleValue{
						Paths: []netv1beta1.HTTPIngressPath{
							{Path: "/"},
						},
					},
				},
			},
			{Host: "blog.presslabs.com"},
		}
		bk = netv1beta1.IngressBackend{}
	})

	When("upserting a new host", func() {
		It("should create a new rule", func() {
			Expect(upsertPath(rules, "docs.presslabs.com", "/", bk)).To(Equal(
				[]netv1beta1.IngressRule{
					{
						Host: "presslabs.com",
						IngressRuleValue: netv1beta1.IngressRuleValue{
							HTTP: &netv1beta1.HTTPIngressRuleValue{
								Paths: []netv1beta1.HTTPIngressPath{
									{Path: "/"},
								},
							},
						},
					},
					{Host: "blog.presslabs.com"},
					{
						Host: "docs.presslabs.com",
						IngressRuleValue: netv1beta1.IngressRuleValue{
							HTTP: &netv1beta1.HTTPIngressRuleValue{
								Paths: []netv1beta1.HTTPIngressPath{
									{Path: "/", Backend: bk},
								},
							},
						},
					},
				}))
		})
	})

	When("upserting an existing host with a new path", func() {
		It("should add the path to the existing rule", func() {
			Expect(upsertPath(rules, "presslabs.com", "/blog", bk)).To(Equal(
				[]netv1beta1.IngressRule{
					{
						Host: "presslabs.com",
						IngressRuleValue: netv1beta1.IngressRuleValue{
							HTTP: &netv1beta1.HTTPIngressRuleValue{
								Paths: []netv1beta1.HTTPIngressPath{
									{Path: "/"},
									{Path: "/blog", Backend: bk},
								},
							},
						},
					},
					{Host: "blog.presslabs.com"},
				}))
		})
	})

	When("upserting an existing host with an existing path", func() {
		It("should not change anything", func() {
			Expect(upsertPath(rules, "presslabs.com", "/", bk)).To(Equal(
				[]netv1beta1.IngressRule{
					{
						Host: "presslabs.com",
						IngressRuleValue: netv1beta1.IngressRuleValue{
							HTTP: &netv1beta1.HTTPIngressRuleValue{
								Paths: []netv1beta1.HTTPIngressPath{
									{Path: "/", Backend: bk},
								},
							},
						},
					},
					{Host: "blog.presslabs.com"},
				}))

		})
	})
})
