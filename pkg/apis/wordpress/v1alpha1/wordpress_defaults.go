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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

func (wp *Wordpress) WithDefaults() (d *Wordpress) {
	d = wp
	if len(d.Spec.VolumeMountsSpec) == 0 {
		d.Spec.VolumeMountsSpec = []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      "content",
				MountPath: "/var/www/html/wp-content",
				SubPath:   "wp-content",
			},
		}
		if d.Spec.MediaVolumeSpec != nil {
			d.Spec.VolumeMountsSpec = append(d.Spec.VolumeMountsSpec, corev1.VolumeMount{
				Name:      "media",
				MountPath: "/var/www/html/wp-content/uploads",
			})
		}
	}
	return
}
