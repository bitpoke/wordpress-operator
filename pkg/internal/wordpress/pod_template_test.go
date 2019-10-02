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

package wordpress

import (
	"fmt"
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	"github.com/presslabs/wordpress-operator/pkg/cmd/options"
)

var _ = Describe("Web pod spec", func() {
	var (
		wp *Wordpress
	)

	BeforeEach(func() {
		name := fmt.Sprintf("cluster-%d", rand.Int31())
		ns := "default"

		wp = New(&wordpressv1alpha1.Wordpress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: ns,
				Labels:    map[string]string{"app.kubernetes.io/part-of": "test"},
			},
			Spec: wordpressv1alpha1.WordpressSpec{
				Domains: []wordpressv1alpha1.Domain{"test.com"},
			},
		})
		wp.SetDefaults()
	})

	DescribeTable("Shouldn't generate any new init containers if git is not configured",
		func(f func() func() corev1.PodTemplateSpec) {
			// we need this hack to allow wp to be initialized
			podSpec := f()
			Expect(podSpec().Spec.InitContainers).To(HaveLen(0))
		},
		Entry("for web pod", func() func() corev1.PodTemplateSpec { return wp.WebPodTemplateSpec }),
		Entry("for job pod", func() func() corev1.PodTemplateSpec {
			return func() corev1.PodTemplateSpec { return wp.JobPodTemplateSpec() }
		}),
	)

	DescribeTable("Should generate a git container when git is configured",
		func(f func() (func() corev1.PodTemplateSpec, *Wordpress)) {
			// we need this hack to allow wp to be initialized with our custom values
			podSpec, w := f()

			w.Spec.CodeVolumeSpec = &wordpressv1alpha1.CodeVolumeSpec{
				GitDir: &wordpressv1alpha1.GitVolumeSource{},
			}
			containers := podSpec().Spec.InitContainers

			Expect(containers).To(HaveLen(1))
			Expect(containers[0].Name).To(Equal("git"))
			Expect(containers[0].Image).To(Equal(options.GitCloneImage))
		},
		Entry("for web pod", func() (func() corev1.PodTemplateSpec, *Wordpress) {
			return wp.WebPodTemplateSpec, wp
		}),
		Entry("for job pod", func() (func() corev1.PodTemplateSpec, *Wordpress) {
			return func() corev1.PodTemplateSpec { return wp.JobPodTemplateSpec("test") }, wp
		}),
	)

	DescribeTable("Should generate an init contaniner, used to install Wordpress",
		func(f func() (func() corev1.PodTemplateSpec, *Wordpress)) {
			// we need this hack to allow wp to be initialized with our custom values
			podSpec, w := f()

			w.Spec.WordpressBootstrapSpec = &wordpressv1alpha1.WordpressBootstrapSpec{}
			containers := podSpec().Spec.InitContainers

			Expect(containers).To(HaveLen(1))
			Expect(containers[0].Name).To(Equal("install-wp"))
			Expect(containers[0].Image).To(Equal(w.Spec.Image))
		},
		Entry("for web pod", func() (func() corev1.PodTemplateSpec, *Wordpress) {
			return wp.WebPodTemplateSpec, wp
		}),
		Entry("for job pod", func() (func() corev1.PodTemplateSpec, *Wordpress) {
			return func() corev1.PodTemplateSpec { return wp.JobPodTemplateSpec("test") }, wp
		}),
	)

	It("should give me the default domain", func() {
		Expect(wp.MainDomain()).To(Equal(wordpressv1alpha1.Domain("test.com")))
		wp.Spec.Domains = []wordpressv1alpha1.Domain{}

		Expect(wp.MainDomain()).To(Equal(wordpressv1alpha1.Domain(fmt.Sprintf("%s.default.svc.cluster.local", wp.Name))))
	})

	It("should give me right home URL", func() {
		Expect(wp.HomeURL()).To(Equal("http://test.com"))
	})
})
