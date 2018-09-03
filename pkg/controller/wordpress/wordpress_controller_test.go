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
	. "github.com/onsi/gomega"

	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
)

const timeout = time.Second * 2

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

	Describe("when creating a new Wordpress resource", func() {
		var expectedRequest reconcile.Request
		var dependantKey types.NamespacedName
		var wp *wordpressv1alpha1.Wordpress
		var rt *wordpressv1alpha1.WordpressRuntime

		BeforeEach(func() {
			r := rand.Int31()
			name := fmt.Sprintf("wp-%d", r)
			runtimeName := fmt.Sprintf("rt-%d", r)
			domain := wordpressv1alpha1.Domain(fmt.Sprintf("%s.example.com", name))

			expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: "default"}}
			dependantKey = types.NamespacedName{Name: name, Namespace: "default"}
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
					Runtime: rt.Name,
					Domains: []wordpressv1alpha1.Domain{domain},
				},
			}

			Expect(c.Create(context.TODO(), rt)).To(Succeed())
		})

		AfterEach(func() {
			// nolint: errcheck
			c.Delete(context.TODO(), rt)
		})

		It("reconciles the deployment", func() {
			// Create the Wordpress object and expect the Reconcile and Deployment to be created
			Expect(c.Create(context.TODO(), wp)).To(Succeed())
			// nolint: errcheck
			defer c.Delete(context.TODO(), wp)

			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

			deploy := &appsv1.Deployment{}
			Eventually(func() error { return c.Get(context.TODO(), dependantKey, deploy) }, timeout).Should(Succeed())

			// Delete the Deployment and expect Reconcile to be called for Deployment deletion
			When("deleting the deployment", func() {
				Expect(c.Delete(context.TODO(), deploy)).To(Succeed())
				Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
				Eventually(func() error { return c.Get(context.TODO(), dependantKey, deploy) }, timeout).Should(Succeed())
			})

			// Manually delete Deployment since GC isn't enabled in the test control plane
			Eventually(func() error { return c.Delete(context.TODO(), deploy) }, timeout).Should(Succeed())
		})

		It("reconciles the service", func() {
			// Create the Wordpress object and expect the Reconcile and Service to be created
			Expect(c.Create(context.TODO(), wp)).To(Succeed())
			// nolint: errcheck
			defer c.Delete(context.TODO(), wp)

			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

			service := &corev1.Service{}
			Eventually(func() error { return c.Get(context.TODO(), dependantKey, service) }, timeout).Should(Succeed())

			// Delete the Service and expect Reconcile to be called for Service deletion
			When("deleting the service", func() {
				Expect(c.Delete(context.TODO(), service)).To(Succeed())
				Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
				Eventually(func() error { return c.Get(context.TODO(), dependantKey, service) }, timeout).Should(Succeed())
			})

			// Manually delete Service since GC isn't enabled in the test control plane
			Eventually(func() error { return c.Delete(context.TODO(), service) }, timeout).Should(Succeed())
		})

		It("reconciles the ingress", func() {
			// Create the Wordpress object and expect the Reconcile and ingress to be created
			Expect(c.Create(context.TODO(), wp)).To(Succeed())
			// nolint: errcheck
			defer c.Delete(context.TODO(), wp)

			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

			ingress := &extv1beta1.Ingress{}
			Eventually(func() error { return c.Get(context.TODO(), dependantKey, ingress) }, timeout).Should(Succeed())

			// Delete the ingress and expect Reconcile to be called for ingress deletion
			When("deleting the ingress", func() {
				Expect(c.Delete(context.TODO(), ingress)).To(Succeed())
				Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
				Eventually(func() error { return c.Get(context.TODO(), dependantKey, ingress) }, timeout).Should(Succeed())
			})

			// Manually delete ingress since GC isn't enabled in the test control plane
			Eventually(func() error { return c.Delete(context.TODO(), ingress) }, timeout).Should(Succeed())
		})

		It("reconciles the webroot pvc", func() {
			dependantKey.Name = fmt.Sprintf("%s-webroot", wp.Name)
			wp.Spec.WebrootVolumeSpec = &wordpressv1alpha1.WordpressVolumeSpec{}
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

			// Create the Wordpress object and expect the Reconcile and pvc to be created
			Expect(c.Create(context.TODO(), wp)).To(Succeed())
			// nolint: errcheck
			defer c.Delete(context.TODO(), wp)

			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

			pvc := &corev1.PersistentVolumeClaim{}
			Eventually(func() error { return c.Get(context.TODO(), dependantKey, pvc) }, timeout).Should(Succeed())

			// Delete the pvc and expect Reconcile to be called for pvc deletion
			When("deleting the pvc", func() {
				Expect(c.Delete(context.TODO(), pvc)).To(Succeed())
				Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
				Eventually(func() error { return c.Get(context.TODO(), dependantKey, pvc) }, timeout).Should(Succeed())
			})

			// Manually delete pvc since GC isn't enabled in the test control plane
			Eventually(func() error { return c.Delete(context.TODO(), pvc) }, timeout).Should(Succeed())
		})

		It("reconciles the media pvc", func() {
			dependantKey.Name = fmt.Sprintf("%s-media", wp.Name)
			wp.Spec.MediaVolumeSpec = &wordpressv1alpha1.WordpressVolumeSpec{}
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

			// Create the Wordpress object and expect the Reconcile and pvc to be created
			Expect(c.Create(context.TODO(), wp)).To(Succeed())
			// nolint: errcheck
			defer c.Delete(context.TODO(), wp)

			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

			pvc := &corev1.PersistentVolumeClaim{}
			Eventually(func() error { return c.Get(context.TODO(), dependantKey, pvc) }, timeout).Should(Succeed())

			// Delete the pvc and expect Reconcile to be called for pvc deletion
			When("deleting the pvc", func() {
				Expect(c.Delete(context.TODO(), pvc)).To(Succeed())
				Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
				Eventually(func() error { return c.Get(context.TODO(), dependantKey, pvc) }, timeout).Should(Succeed())
			})

			// Manually delete pvc since GC isn't enabled in the test control plane
			Eventually(func() error { return c.Delete(context.TODO(), pvc) }, timeout).Should(Succeed())
		})

		It("reconciles the wp-cron", func() {
			// Create the Wordpress object and expect the Reconcile and WPCron to be created
			Expect(c.Create(context.TODO(), wp)).To(Succeed())
			// nolint: errcheck
			defer c.Delete(context.TODO(), wp)

			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

			cron := &batchv1beta1.CronJob{}
			dependantKey := types.NamespacedName{
				Name:      wp.GetWPCronName(),
				Namespace: wp.Namespace,
			}
			Eventually(func() error { return c.Get(context.TODO(), dependantKey, cron) }, timeout).Should(Succeed())

			// Delete the WPCron and expect Reconcile to be called for WPCron deletion
			When("deleting the wp-cron", func() {
				Expect(c.Delete(context.TODO(), cron)).To(Succeed())
				Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
				Eventually(func() error { return c.Get(context.TODO(), dependantKey, cron) }, timeout).Should(Succeed())
			})

			// Manually delete WPCron since GC isn't enabled in the test control plane
			Eventually(func() error { return c.Delete(context.TODO(), cron) }, timeout).Should(Succeed())
		})
	})
})
