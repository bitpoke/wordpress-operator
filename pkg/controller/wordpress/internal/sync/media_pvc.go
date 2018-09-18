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
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/presslabs/controller-util/syncer"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	"github.com/presslabs/wordpress-operator/pkg/controller/internal/wordpress"
)

func getMediaPVCName(wp *wordpressv1alpha1.Wordpress) string { return fmt.Sprintf("%s-media", wp.Name) }

// NewMediaPVCSyncer returns a new sync.Interface for reconciling media PVC
func NewMediaPVCSyncer(wp *wordpressv1alpha1.Wordpress, rt *wordpressv1alpha1.WordpressRuntime, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	obj := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getMediaPVCName(wp),
			Namespace: wp.Namespace,
		},
	}
	return syncer.NewObjectSyncer("MediaPVC", wp, obj, c, scheme, func(existing runtime.Object) error {
		out := existing.(*corev1.PersistentVolumeClaim)

		out.Labels = wordpress.LabelsForTier(wp, "front")

		// PVC spec is immutable
		if !reflect.DeepEqual(out.Spec, corev1.PersistentVolumeClaimSpec{}) {
			return nil
		}

		volSpec := rt.Spec.MediaVolumeSpec
		if wp.Spec.MediaVolumeSpec != nil {
			volSpec = wp.Spec.MediaVolumeSpec
		}

		out.Spec = *volSpec.PersistentVolumeClaim

		return nil
	})
}
