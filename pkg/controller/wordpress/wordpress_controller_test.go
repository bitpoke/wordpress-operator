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

package wordpress

import (
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
)

const timeout = time.Second * 5

var _ = Describe("Wordpress controller", func() {
	var (
		// channel for incoming reconcile requests
		requests chan reconcile.Request
		// stop channel for controller manager
		stop chan struct{}
		// controller k8s client
		c client.Client
	)

	BeforeEach(func() {
		var recFn reconcile.Reconciler

		mgr, err := manager.New(cfg, manager.Options{})
		Expect(err).NotTo(HaveOccurred())
		c = mgr.GetClient()

		recFn, requests = SetupTestReconcile(newReconciler(mgr))
		Expect(add(mgr, recFn)).To(Succeed())

		stop = StartTestManager(mgr)
	})

	AfterEach(func() {
		close(stop)
	})

	When("creating a new Wordpress resource", func() {
		var (
			expectedRequest reconcile.Request
			wp              *wordpressv1alpha1.Wordpress
			rt              *wordpressv1alpha1.WordpressRuntime
		)

		BeforeEach(func() {
			r := rand.Int31()
			name := fmt.Sprintf("wp-%d", r)
			runtimeName := fmt.Sprintf("rt-%d", r)
			domain := wordpressv1alpha1.Domain(fmt.Sprintf("%s.example.com", name))
			expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: "default"}}

			rt = &wordpressv1alpha1.WordpressRuntime{
				ObjectMeta: metav1.ObjectMeta{Name: runtimeName},
				Spec: wordpressv1alpha1.WordpressRuntimeSpec{
					DefaultImage: "docker.io/library/hello-world",
					WebPodTemplate: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "wordpress",
									Image: "image",
									Ports: []corev1.ContainerPort{
										{
											Name:          "http",
											ContainerPort: 80,
											Protocol:      corev1.ProtocolTCP,
										},
									},
								},
							},
						},
					},
					CLIPodTemplate: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "wp-cli",
									Image: "cli-image",
								},
							},
						},
					},
				},
			}
			wp = &wordpressv1alpha1.Wordpress{
				ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
				Spec: wordpressv1alpha1.WordpressSpec{
					Runtime:           rt.Name,
					Domains:           []wordpressv1alpha1.Domain{domain},
					WebrootVolumeSpec: &wordpressv1alpha1.WordpressVolumeSpec{},
					MediaVolumeSpec:   &wordpressv1alpha1.WordpressVolumeSpec{},
				},
			}

			wp.Spec.WebrootVolumeSpec.PersistentVolumeClaim = &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("1Gi"),
					},
				},
			}

			wp.Spec.MediaVolumeSpec.PersistentVolumeClaim = &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("1Gi"),
					},
				},
			}

			Expect(c.Create(context.TODO(), rt)).To(Succeed())
			Expect(c.Create(context.TODO(), wp)).To(Succeed())
			// We should receive the initial reconciliation request
			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

			// We need to drain the requests queue because syncing a subresource
			// might trigger reconciliation again and we want to isolate tests
			// to their own reconciliation requests
			done := time.After(time.Second)
			for {
				select {
				case <-requests:
					continue
				case <-done:
					return
				}
			}
		})

		// nolint: errcheck
		AfterEach(func() {
			Expect(c.Delete(context.TODO(), wp)).To(Succeed())
			Expect(c.Delete(context.TODO(), rt)).To(Succeed())
		})

		DescribeTable("the reconciler",
			func(nameFmt string, obj runtime.Object) {
				key := types.NamespacedName{
					Name:      fmt.Sprintf(nameFmt, wp.Name),
					Namespace: wp.Namespace,
				}
				Eventually(func() error { return c.Get(context.TODO(), key, obj) }, timeout).Should(Succeed())

				// Delete the resource and expect Reconcile to be called
				Expect(c.Delete(context.TODO(), obj)).To(Succeed())
				Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
				Eventually(func() error { return c.Get(context.TODO(), key, obj) }, timeout).Should(Succeed())
			},
			Entry("reconciles the deployment", "%s", &appsv1.Deployment{}),
			Entry("reconciles the service", "%s", &corev1.Service{}),
			Entry("reconciles the ingress", "%s", &extv1beta1.Ingress{}),
			Entry("reconciles the webroot pvc", "%s-webroot", &corev1.PersistentVolumeClaim{}),
			Entry("reconciles the media pvc", "%s-media", &corev1.PersistentVolumeClaim{}),
			Entry("reconciles the wp-cron", "%s-wp-cron", &batchv1beta1.CronJob{}),
		)
	})
})
