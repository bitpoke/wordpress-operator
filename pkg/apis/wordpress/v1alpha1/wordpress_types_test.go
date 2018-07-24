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

package v1alpha1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"golang.org/x/net/context"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
)

var _ = Describe("Wordpress CRUD", func() {
	var created *Wordpress
	var key types.NamespacedName

	BeforeEach(func() {
		key = types.NamespacedName{Name: "foo", Namespace: "default"}

		created = &Wordpress{}
		created.Name = key.Name
		created.Namespace = key.Namespace
		created.Spec.Domains = []Domain{"example.com"}
	})

	AfterEach(func() {
		c.Delete(context.TODO(), created)
	})

	Describe("when sending a storage request", func() {
		Context("for a valid config", func() {
			It("should provide CRUD access to the object", func() {
				fetched := &Wordpress{}

				By("returning success from the create request")
				Expect(c.Create(context.TODO(), created)).Should(Succeed())

				By("returning the same object as created")
				Expect(c.Get(context.TODO(), key, fetched)).Should(Succeed())
				Expect(fetched).To(Equal(created))

				By("allowing label updates")
				updated := fetched.DeepCopy()
				updated.Labels = map[string]string{"hello": "world"}
				Expect(c.Update(context.TODO(), updated)).Should(Succeed())
				Expect(c.Get(context.TODO(), key, fetched)).Should(Succeed())
				Expect(fetched).To(Equal(updated))

				By("deleting an fetched object")
				Expect(c.Delete(context.TODO(), fetched)).Should(Succeed())
				Expect(c.Get(context.TODO(), key, fetched)).To(HaveOccurred())
			})
		})
	})
})
