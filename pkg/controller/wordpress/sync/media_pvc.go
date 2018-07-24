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
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
)

const (
	EventReasonMediaPVCFailed  EventReason = "MediaPVCFailed"
	EventReasonMediaPVCUpdated EventReason = "MediaPVCUpdated"
)

type MediaPVCSyncer struct {
	scheme   *runtime.Scheme
	wp       *wordpressv1alpha1.Wordpress
	key      types.NamespacedName
	existing *corev1.PersistentVolumeClaim
}

var _ Interface = &MediaPVCSyncer{}

func NewMediaPVCSyncer(wp *wordpressv1alpha1.Wordpress, r *runtime.Scheme) *MediaPVCSyncer {
	return &MediaPVCSyncer{
		scheme:   r,
		wp:       wp,
		existing: &corev1.PersistentVolumeClaim{},
		key: types.NamespacedName{
			Name:      wp.GetMediaPVCName(),
			Namespace: wp.Namespace,
		},
	}
}

func (s *MediaPVCSyncer) GetKey() types.NamespacedName                 { return s.key }
func (s *MediaPVCSyncer) GetExistingObjectPlaceholder() runtime.Object { return s.existing }

func (s *MediaPVCSyncer) T(in runtime.Object) (runtime.Object, error) {
	out := in.(*corev1.PersistentVolumeClaim)

	out.Name = s.key.Name
	out.Namespace = s.key.Namespace
	out.Labels = s.wp.LabelsForTier("front")
	if err := controllerutil.SetControllerReference(s.wp, out, s.scheme); err != nil {
		return nil, err
	}

	// PVC spec is immutable
	if !reflect.DeepEqual(out.Spec, corev1.PersistentVolumeClaimSpec{}) {
		return out, nil
	}

	out.Spec = *s.wp.Spec.MediaVolumeSpec.PersistentVolumeClaim

	return out, nil
}

func (s *MediaPVCSyncer) GetErrorEventReason(err error) EventReason {
	if err == nil {
		return EventReasonMediaPVCUpdated
	} else {
		return EventReasonMediaPVCFailed
	}
}
