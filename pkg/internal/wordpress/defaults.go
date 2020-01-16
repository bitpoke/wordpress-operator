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

	"path"

	"github.com/presslabs/wordpress-operator/pkg/cmd/options"
)

const (
	codeSrcMountPath = "/var/run/presslabs.org/code/src"

	defaultCodeMountPath   = "/app/web/wp-content"
	defaultRepoCodeSubPath = "wp-content"

	configMountPath          = "/app/config"
	defaultRepoConfigSubPath = "config"

	mediaSubPath          = "uploads"
	defaultMediaMountPath = defaultCodeMountPath + "/" + mediaSubPath
)

// SetDefaults sets Wordpress field defaults
// nolint: gocyclo
func (wp *Wordpress) SetDefaults() {
	if len(wp.Spec.Image) == 0 {
		wp.Spec.Image = options.WordpressRuntimeImage
	}

	if len(wp.Spec.ImagePullPolicy) == 0 {
		wp.Spec.ImagePullPolicy = corev1.PullAlways
	}

	if wp.Spec.CodeVolumeSpec != nil && wp.Spec.CodeVolumeSpec.MountPath == "" {
		wp.Spec.CodeVolumeSpec.MountPath = defaultCodeMountPath
	}

	if wp.Spec.CodeVolumeSpec != nil && wp.Spec.CodeVolumeSpec.ContentSubPath == "" {
		wp.Spec.CodeVolumeSpec.ContentSubPath = defaultRepoCodeSubPath
	}

	if wp.Spec.CodeVolumeSpec != nil && wp.Spec.CodeVolumeSpec.ConfigSubPath == "" {
		wp.Spec.CodeVolumeSpec.ConfigSubPath = defaultRepoConfigSubPath
	}

	if wp.Spec.MediaVolumeSpec != nil && wp.Spec.MediaVolumeSpec.MountPath == "" {
		if wp.Spec.CodeVolumeSpec != nil {
			wp.Spec.MediaVolumeSpec.MountPath = path.Join(wp.Spec.MediaVolumeSpec.MountPath, mediaSubPath)
		} else {
			wp.Spec.MediaVolumeSpec.MountPath = defaultMediaMountPath
		}
	}
}
