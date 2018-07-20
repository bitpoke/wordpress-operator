/*
Copyright 2018 Pressinfra SRL

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
	"reflect"

	"github.com/golang/glog"

	core_util "github.com/appscode/kutil/core/v1"
	corev1 "k8s.io/api/core/v1"

	wpapi "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	"github.com/presslabs/wordpress-operator/pkg/factory/wordpress"
)

func (c *Controller) syncPVC(wp *wpapi.Wordpress) error {
	wpf := wordpress.Generator{WP: wp}
	labels := wpf.Labels()

	if wp.Spec.ContentVolumeSpec.PersistentVolumeClaim != nil {
		meta := c.objectMeta(wp, wpf.ContentPVCName())
		meta.Labels = labels
		_, _, err := core_util.CreateOrPatchPVC(c.KubeClient, meta, func(in *corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaim {
			if wp.Spec.ContentVolumeSpec.PersistentVolumeClaim != nil {
				if reflect.DeepEqual(in.Spec, corev1.PersistentVolumeClaimSpec{}) {
					in.Spec = *wp.Spec.ContentVolumeSpec.PersistentVolumeClaim
				} else {
					glog.V(4).Infof("Skip updating PersistentVolumeClaim %s/%s. The PersistentVolumeClaim spec is immutable after creation.", wp.ObjectMeta.Namespace, wp.ObjectMeta.Name)
				}
			}
			return in
		})
		if err != nil {
			return err
		}
	}

	if wp.Spec.MediaVolumeSpec == nil {
		// Quick return if MediaVolumeSpec is not defined
		return nil
	}
	if wp.Spec.MediaVolumeSpec.PersistentVolumeClaim != nil {
		meta := c.objectMeta(wp, wpf.MediaPVCName())
		meta.Labels = labels
		_, _, err := core_util.CreateOrPatchPVC(c.KubeClient, meta, func(in *corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaim {
			if wp.Spec.MediaVolumeSpec.PersistentVolumeClaim != nil {
				if reflect.DeepEqual(in.Spec, corev1.PersistentVolumeClaimSpec{}) {
					in.Spec = *wp.Spec.MediaVolumeSpec.PersistentVolumeClaim
				} else {
					glog.V(4).Infof("Skip updating PersistentVolumeClaim %s/%s. The PersistentVolumeClaim spec is immutable after creation.", wp.ObjectMeta.Namespace, wp.ObjectMeta.Name)
				}
			}
			return in
		})
		if err != nil {
			return err
		}
	}
	return nil
}
