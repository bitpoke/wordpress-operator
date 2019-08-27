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
	corev1 "k8s.io/api/core/v1"

	"github.com/presslabs/wordpress-operator/pkg/cmd/options"
)

const (
	codeSrcMountPath = "/var/run/presslabs.org/code/src"

	defaultCodeMountPath   = "/app/web/wp-content"
	defaultRepoCodeSubPath = "wp-content"

	configMountPath          = "/app/config"
	defaultRepoConfigSubPath = "config"
)

// SetDefaults sets Wordpress field defaults
func (o *Wordpress) SetDefaults() {
	if len(o.Spec.Image) == 0 {
		o.Spec.Image = options.WordpressRuntimeImage
	}

	if len(o.Spec.ImagePullPolicy) == 0 {
		o.Spec.ImagePullPolicy = corev1.PullAlways
	}

	if o.Spec.CodeVolumeSpec != nil && len(o.Spec.CodeVolumeSpec.MountPath) == 0 {
		o.Spec.CodeVolumeSpec.MountPath = defaultCodeMountPath
	}

	if o.Spec.CodeVolumeSpec != nil && len(o.Spec.CodeVolumeSpec.ContentSubPath) == 0 {
		o.Spec.CodeVolumeSpec.ContentSubPath = defaultRepoCodeSubPath
	}

	if o.Spec.CodeVolumeSpec != nil && len(o.Spec.CodeVolumeSpec.ConfigSubPath) == 0 {
		o.Spec.CodeVolumeSpec.ConfigSubPath = defaultRepoConfigSubPath
	}
}
