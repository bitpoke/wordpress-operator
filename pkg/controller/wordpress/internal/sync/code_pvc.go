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
	"errors"
	"reflect"

	"github.com/presslabs/controller-util/syncer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/bitpoke/wordpress-operator/pkg/internal/wordpress"
)

var errCodeVolumeClaimNotDefined = errors.New(".spec.code.persistentVolumeClaim is not defined")

// NewCodePVCSyncer returns a new sync.Interface for reconciling codePVC.
func NewCodePVCSyncer(wp *wordpress.Wordpress, c client.Client) syncer.Interface {
	objLabels := wp.ComponentLabels(wordpress.WordpressCodePVC)

	obj := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wp.ComponentName(wordpress.WordpressCodePVC),
			Namespace: wp.Namespace,
		},
	}

	return syncer.NewObjectSyncer("CodePVC", wp.Unwrap(), obj, c, func() error {
		obj.Labels = labels.Merge(labels.Merge(wp.Spec.CodeVolumeSpec.Labels, objLabels), controllerLabels)

		if len(wp.Spec.CodeVolumeSpec.Annotations) > 0 {
			obj.Annotations = labels.Merge(obj.Annotations, wp.Spec.CodeVolumeSpec.Annotations)
		}

		if wp.Spec.CodeVolumeSpec == nil || wp.Spec.CodeVolumeSpec.PersistentVolumeClaim == nil {
			return errCodeVolumeClaimNotDefined
		}

		// PVC spec is immutable
		if !reflect.DeepEqual(obj.Spec, corev1.PersistentVolumeClaimSpec{}) {
			return nil
		}

		obj.Spec = *wp.Spec.CodeVolumeSpec.PersistentVolumeClaim

		return nil
	})
}
