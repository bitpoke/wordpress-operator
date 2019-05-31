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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"math/rand"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	DescribeTable("Shouldn't generate any init containers, if git and a remote stream is not configured",
		func(f func() func() corev1.PodTemplateSpec) {
			// we need this hack to allow wp to be initialized
			podSpec := f()
			Expect(podSpec().Spec.InitContainers).To(BeEmpty())
		},
		Entry("for web pod", func() func() corev1.PodTemplateSpec { return wp.WebPodTemplateSpec }),
		Entry("for job pod", func() func() corev1.PodTemplateSpec {
			return func() corev1.PodTemplateSpec { return wp.JobPodTemplateSpec() }
		}),
	)

	DescribeTable("Should generate just a git container, when no remote stream is configured",
		func(f func() (func() corev1.PodTemplateSpec, *Wordpress)) {
			// we need this hack to allow wp to be initialized with our custom values
			podSpec, w := f()

			w.Spec.CodeVolumeSpec = &wordpressv1alpha1.CodeVolumeSpec{
				GitDir: &wordpressv1alpha1.GitVolumeSource{},
			}
			containers := podSpec().Spec.InitContainers

			Expect(containers).To(HaveLen(1))
			Expect(containers[0].Name).To(Equal("git"))
			Expect(containers[0].Image).To(Equal(gitCloneImage))
		},
		Entry("for web pod", func() (func() corev1.PodTemplateSpec, *Wordpress) {
			return wp.WebPodTemplateSpec, wp
		}),
		Entry("for job pod", func() (func() corev1.PodTemplateSpec, *Wordpress) {
			return func() corev1.PodTemplateSpec { return wp.JobPodTemplateSpec("test") }, wp
		}),
	)

	DescribeTable("Should generate a git container and a rclone container when gcs is configured",
		func(f func() (func() corev1.PodTemplateSpec, *Wordpress)) {
			podSpec, w := f()

			w.Spec.CodeVolumeSpec = &wordpressv1alpha1.CodeVolumeSpec{
				GitDir: &wordpressv1alpha1.GitVolumeSource{},
			}
			w.Spec.MediaVolumeSpec = &wordpressv1alpha1.MediaVolumeSpec{
				GCSVolumeSource: &wordpressv1alpha1.GCSVolumeSource{Env: []corev1.EnvVar{}},
			}

			spec := podSpec()

			initContainers := spec.Spec.InitContainers
			Expect(initContainers).To(HaveLen(2))

			Expect(initContainers[0].Name).To(Equal("rclone-init-ftp"))
			Expect(initContainers[0].Image).To(Equal(rcloneImage))

			Expect(initContainers[1].Name).To(Equal("git"))
			Expect(initContainers[1].Image).To(Equal(gitCloneImage))

			containers := spec.Spec.Containers
			Expect(containers).To(HaveLen(2))

			Expect(containers[1].Name).To(Equal("rclone-ftp"))
			Expect(containers[1].Image).To(Equal(rcloneImage))
			Expect(containers[1].Args).To(Equal(
				[]string{"serve", "ftp", "-vvv", "--vfs-cache-max-age", "30s", "--vfs-cache-mode", "full",
					"--vfs-cache-poll-interval", "0", "--poll-interval", "0", "$(RCLONE_STREAM)/",
					fmt.Sprintf("--addr=0.0.0.0:%d", mediaFTPPort),
				}))
			Expect(containers[1].Env).To(Equal([]corev1.EnvVar{
				{
					Name:  "RCLONE_STREAM",
					Value: "gs:",
				},
			}))
		},

		Entry("for web pod", func() (func() corev1.PodTemplateSpec, *Wordpress) {
			return wp.WebPodTemplateSpec, wp
		}),
		Entry("for job pod", func() (func() corev1.PodTemplateSpec, *Wordpress) {
			return func() corev1.PodTemplateSpec { return wp.JobPodTemplateSpec("test") }, wp
		}),
	)

	DescribeTable("Should create an install wp init container container if the credentials are in "+
		"the environment",
		func(f func() (func() corev1.PodTemplateSpec, *Wordpress)) {
			// we need this hack to allow wp to be initialized with our custom values
			podSpec, w := f()

			w.Spec.WordpressBootstrapSpec = &wordpressv1alpha1.WordpressBootstrapSpec{
				Env: []corev1.EnvVar{
					{
						Name:  "WORDPRESS_BOOTSTRAP_USER",
						Value: "test",
					},
					{
						Name:  "WORDPRESS_BOOTSTRAP_PASSWORD",
						Value: "test",
					},
				},
			}

			initContainers := podSpec().Spec.InitContainers
			Expect(initContainers).To(HaveLen(1))

			Expect(initContainers[0].Name).To(Equal("install-wp"))
			Expect(initContainers[0].Image).To(Equal(w.image()))
		},
		Entry("for web pod", func() (func() corev1.PodTemplateSpec, *Wordpress) {
			return wp.WebPodTemplateSpec, wp
		}),
		Entry("for job pod", func() (func() corev1.PodTemplateSpec, *Wordpress) {
			return func() corev1.PodTemplateSpec { return wp.JobPodTemplateSpec("test") }, wp
		}),
	)
})
