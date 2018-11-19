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
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/presslabs/controller-util/syncer"

	"github.com/presslabs/wordpress-operator/pkg/internal/wordpress"
)

// NewCodePVCSyncer returns a new sync.Interface for reconciling codePVC
func NewCodePVCSyncer(wp *wordpress.Wordpress, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	objLabels := wp.ComponentLabels(wordpress.WordpressCodePVC)

	obj := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wp.ComponentName(wordpress.WordpressCodePVC),
			Namespace: wp.Namespace,
		},
	}
	return syncer.NewObjectSyncer("CodePVC", wp.Unwrap(), obj, c, scheme, func(existing runtime.Object) error {
		out := existing.(*corev1.PersistentVolumeClaim)
		out.Labels = labels.Merge(labels.Merge(out.Labels, objLabels), controllerLabels)

		if wp.Spec.CodeVolumeSpec == nil || wp.Spec.CodeVolumeSpec.PersistentVolumeClaim == nil {
			return fmt.Errorf(".spec.code.persistentVolumeClaim is not defined")
		}

		// PVC spec is immutable
		if !reflect.DeepEqual(out.Spec, corev1.PersistentVolumeClaimSpec{}) {
			return nil
		}

		out.Spec = *wp.Spec.CodeVolumeSpec.PersistentVolumeClaim

		return nil
	})
}
