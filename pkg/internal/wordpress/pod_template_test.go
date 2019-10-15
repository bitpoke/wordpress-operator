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
				Routes: []wordpressv1alpha1.RouteSpec{
					{
						Domain: "test.com",
					},
				},
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

	It("should generate a valid STACK_ROUTES", func() {
		spec := wp.WebPodTemplateSpec()
		e, found := lookupEnvVar("STACK_ROUTES", spec.Spec.Containers[0].Env)
		Expect(found).To(BeTrue())
		Expect(e.Value).To(Equal("test.com"))
	})

	It("should generate a valid STACK_ROUTES when routes list is empty", func() {
		wp.Spec.Routes = []wordpressv1alpha1.RouteSpec{}
		spec := wp.WebPodTemplateSpec()
		e, found := lookupEnvVar("STACK_ROUTES", spec.Spec.Containers[0].Env)
		Expect(found).To(BeTrue())
		Expect(e.Value).To(Equal(fmt.Sprintf("%s.default.svc", wp.Name)))
	})

	It("should generate a valid STACK_ROUTES keeping routes order", func() {
		wp.Spec.Routes = []wordpressv1alpha1.RouteSpec{
			{
				Domain: "test.com",
			},
			{
				Domain: "test.org",
			},
		}
		spec := wp.WebPodTemplateSpec()
		e, found := lookupEnvVar("STACK_ROUTES", spec.Spec.Containers[0].Env)
		Expect(found).To(BeTrue())
		Expect(e.Value).To(Equal("test.com,test.org"))
	})

	It("should generate a valid STACK_ROUTES with paths w/o trailing slash", func() {
		wp.Spec.Routes = []wordpressv1alpha1.RouteSpec{
			{
				Domain: "test.com",
				Path:   "/",
			},
			{
				Domain: "test.org",
				Path:   "/abc",
			},
			{
				Domain: "test.net",
				Path:   "/xyz/",
			},
		}
		spec := wp.WebPodTemplateSpec()
		e, found := lookupEnvVar("STACK_ROUTES", spec.Spec.Containers[0].Env)
		Expect(found).To(BeTrue())
		Expect(e.Value).To(Equal("test.com,test.org/abc,test.net/xyz"))
	})

	It("should give me the default domain", func() {
		Expect(wp.MainDomain()).To(Equal("test.com"))

		wp.Spec.Routes = []wordpressv1alpha1.RouteSpec{}
		Expect(wp.MainDomain()).To(Equal(fmt.Sprintf("%s.default.svc", wp.Name)))
	})

	It("should give me right home URL, without trailing slash", func() {
		// WP_HOME and WP_SITEURL should not contain a trailing slash,
		// as per: https://wordpress.org/support/article/changing-the-site-url/

		Expect(wp.HomeURL()).To(Equal("http://test.com"))
	})

	It("should give me right home URL for subpath", func() {
		Expect(wp.HomeURL("foo")).To(Equal("http://test.com/foo"))
		Expect(wp.HomeURL("foo", "bar")).To(Equal("http://test.com/foo/bar"))
		Expect(wp.HomeURL("foo/bar")).To(Equal("http://test.com/foo/bar"))

		wp.Spec.Routes[0].Path = "/subpath"
		Expect(wp.HomeURL()).To(Equal("http://test.com/subpath"))
	})
})

// nolint: unparam
func lookupEnvVar(name string, env []corev1.EnvVar) (corev1.EnvVar, bool) {
	for _, e := range env {
		if e.Name == name {
			return e, true
		}
	}
	return corev1.EnvVar{}, false
}
