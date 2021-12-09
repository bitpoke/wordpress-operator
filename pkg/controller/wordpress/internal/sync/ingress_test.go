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

	netv1 "k8s.io/api/networking/v1"
)

var _ = Describe("The upsertPath function", func() {
	var (
		rules          []netv1.IngressRule
		bk             netv1.IngressBackend
		pathTypePrefix netv1.PathType
	)

	BeforeEach(func() {
		pathTypePrefix = netv1.PathTypePrefix
		rules = []netv1.IngressRule{
			{
				Host: "bitpoke.io",
				IngressRuleValue: netv1.IngressRuleValue{
					HTTP: &netv1.HTTPIngressRuleValue{
						Paths: []netv1.HTTPIngressPath{
							{Path: "/", PathType: &pathTypePrefix},
						},
					},
				},
			},
			{Host: "blog.bitpoke.io"},
		}
		bk = netv1.IngressBackend{}
	})

	When("upserting a new host", func() {
		It("should create a new rule", func() {
			Expect(upsertPath(rules, "docs.bitpoke.io", "/", bk)).To(Equal(
				[]netv1.IngressRule{
					{
						Host: "bitpoke.io",
						IngressRuleValue: netv1.IngressRuleValue{
							HTTP: &netv1.HTTPIngressRuleValue{
								Paths: []netv1.HTTPIngressPath{
									{Path: "/", PathType: &pathTypePrefix},
								},
							},
						},
					},
					{Host: "blog.bitpoke.io"},
					{
						Host: "docs.bitpoke.io",
						IngressRuleValue: netv1.IngressRuleValue{
							HTTP: &netv1.HTTPIngressRuleValue{
								Paths: []netv1.HTTPIngressPath{
									{Path: "/", PathType: &pathTypePrefix, Backend: bk},
								},
							},
						},
					},
				}))
		})
	})

	When("upserting an existing host with a new path", func() {
		It("should add the path to the existing rule", func() {
			Expect(upsertPath(rules, "bitpoke.io", "/blog", bk)).To(Equal(
				[]netv1.IngressRule{
					{
						Host: "bitpoke.io",
						IngressRuleValue: netv1.IngressRuleValue{
							HTTP: &netv1.HTTPIngressRuleValue{
								Paths: []netv1.HTTPIngressPath{
									{Path: "/", PathType: &pathTypePrefix},
									{Path: "/blog", PathType: &pathTypePrefix, Backend: bk},
								},
							},
						},
					},
					{Host: "blog.bitpoke.io"},
				}))
		})
	})

	When("upserting an existing host with an existing path", func() {
		It("should not change anything", func() {
			Expect(upsertPath(rules, "bitpoke.io", "/", bk)).To(Equal(
				[]netv1.IngressRule{
					{
						Host: "bitpoke.io",
						IngressRuleValue: netv1.IngressRuleValue{
							HTTP: &netv1.HTTPIngressRuleValue{
								Paths: []netv1.HTTPIngressPath{
									{Path: "/", PathType: &pathTypePrefix, Backend: bk},
								},
							},
						},
					},
					{Host: "blog.bitpoke.io"},
				}))

		})
	})
})
