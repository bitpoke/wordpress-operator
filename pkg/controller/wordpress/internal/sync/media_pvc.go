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

	"github.com/presslabs/wordpress-operator/pkg/internal/wordpress"
)

var errMediaVolumeClaimNotDefined = errors.New(".spec.media.persistentVolumeClaim is not defined")

// NewMediaPVCSyncer returns a new sync.Interface for reconciling media PVC.
func NewMediaPVCSyncer(wp *wordpress.Wordpress, c client.Client) syncer.Interface {
	objLabels := wp.ComponentLabels(wordpress.WordpressMediaPVC)

	obj := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wp.ComponentName(wordpress.WordpressMediaPVC),
			Namespace: wp.Namespace,
		},
	}

	return syncer.NewObjectSyncer("MediaPVC", wp.Unwrap(), obj, c, func() error {
		obj.Labels = labels.Merge(labels.Merge(wp.Spec.MediaVolumeSpec.Labels, objLabels), controllerLabels)

		if len(wp.Spec.MediaVolumeSpec.Annotations) > 0 {
			obj.Annotations = labels.Merge(obj.Annotations, wp.Spec.MediaVolumeSpec.Annotations)
		}

		if wp.Spec.MediaVolumeSpec == nil || wp.Spec.MediaVolumeSpec.PersistentVolumeClaim == nil {
			return errMediaVolumeClaimNotDefined
		}

		// PVC spec is immutable
		if !reflect.DeepEqual(obj.Spec, corev1.PersistentVolumeClaimSpec{}) {
			return nil
		}

		obj.Spec = *wp.Spec.MediaVolumeSpec.PersistentVolumeClaim

		return nil
	})
}
