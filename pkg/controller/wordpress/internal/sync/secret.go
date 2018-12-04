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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/presslabs/controller-util/rand"
	"github.com/presslabs/controller-util/syncer"

	"github.com/presslabs/wordpress-operator/pkg/internal/wordpress"
)

var (
	generatedSalts = map[string]int{
		"AUTH_KEY":         64,
		"SECURE_AUTH_KEY":  64,
		"LOGGED_IN_KEY":    64,
		"NONCE_KEY":        64,
		"AUTH_SALT":        64,
		"SECURE_AUTH_SALT": 64,
		"LOGGED_IN_SALT":   64,
		"NONCE_SALT":       64,
	}
)

// NewSecretSyncer returns a new sync.Interface for reconciling wordpress secret
func NewSecretSyncer(wp *wordpress.Wordpress, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	objLabels := wp.ComponentLabels(wordpress.WordpressSecret)

	obj := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wp.ComponentName(wordpress.WordpressSecret),
			Namespace: wp.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Secret", wp.Unwrap(), obj, c, scheme, func(existing runtime.Object) error {
		out := existing.(*corev1.Secret)
		out.Labels = labels.Merge(labels.Merge(out.Labels, objLabels), controllerLabels)

		if len(out.Data) == 0 {
			out.Data = make(map[string][]byte)
		}

		for name, size := range generatedSalts {
			if len(out.Data[name]) == 0 {
				random, err := rand.ASCIIString(size)
				if err != nil {
					return err
				}
				out.Data[name] = []byte(random)
			}
		}

		return nil
	})
}
