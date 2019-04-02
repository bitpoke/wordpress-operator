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
	"fmt"
	"path"

	corev1 "k8s.io/api/core/v1"
)

const (
	// InternalHTTPPort represents the internal port used by the runtime container
	InternalHTTPPort = 8080
	gitCloneImage    = "docker.io/library/buildpack-deps:stretch-scm"
	rcloneImage      = "quay.io/presslabs/rclone@sha256:4436a1e2d471236eafac605b24a66f5f18910b6f9cde505db065506208f73f96"
	mediaHTTPPort    = 8090
	mediaFTPPort     = 2121
	codeVolumeName   = "code"
	mediaVolumeName  = "media"
	s3Prefix         = "s3"
	gcsPrefix        = "gs"
)

const gitCloneScript = `#!/bin/bash
set -e
set -o pipefail

export HOME="$(mktemp -d)"
export GIT_SSH_COMMAND="ssh -o UserKnownHostsFile=$HOME/.ssh/knonw_hosts -o StrictHostKeyChecking=no"

test -d "$HOME/.ssh" || mkdir "$HOME/.ssh"

if [ ! -z "$SSH_RSA_PRIVATE_KEY" ] ; then
    echo "$SSH_RSA_PRIVATE_KEY" > "$HOME/.ssh/id_rsa"
    chmod 0400 "$HOME/.ssh/id_rsa"
    export GIT_SSH_COMMAND="$GIT_SSH_COMMAND -o IdentityFile=$HOME/.ssh/id_rsa"
fi

if [ -z "$GIT_CLONE_URL" ] ; then
    echo "No \$GIT_CLONE_URL specified" >&2
    exit 1
fi

find "$SRC_DIR" -maxdepth 1 -mindepth 1 -print0 | xargs -0 /bin/rm -rf

set -x
git clone "$GIT_CLONE_URL" "$SRC_DIR"
cd "$SRC_DIR"
git checkout -B "$GIT_CLONE_REF" "$GIT_CLONE_REF"
`

var (
	wwwDataUserID int64 = 33
)

func (wp *Wordpress) image() string {
	return fmt.Sprintf("%s:%s", wp.Spec.Image, wp.Spec.Tag)
}

func (wp *Wordpress) hasExternalMedia() bool {
	return wp.Spec.MediaVolumeSpec != nil &&
		(wp.Spec.MediaVolumeSpec.S3VolumeSource != nil || wp.Spec.MediaVolumeSpec.GCSVolumeSource != nil)
}

func (wp *Wordpress) env() []corev1.EnvVar {
	scheme := "http"
	if len(wp.Spec.TLSSecretRef) > 0 {
		scheme = "https"
	}

	out := append([]corev1.EnvVar{
		{
			Name:  "WP_HOME",
			Value: fmt.Sprintf("%s://%s", scheme, wp.Spec.Domains[0]),
		},
		{
			Name:  "WP_SITEURL",
			Value: fmt.Sprintf("%s://%s/wp", scheme, wp.Spec.Domains[0]),
		},
	}, wp.Spec.Env...)

	if wp.hasExternalMedia() {
		out = append([]corev1.EnvVar{
			{
				Name:  "UPLOADS_FTP_HOST",
				Value: fmt.Sprintf("127.0.0.1:%d", mediaFTPPort),
			},
			{
				Name:  "UPLOADS_HTTP_PROXY",
				Value: fmt.Sprintf("127.0.0.1:%d", mediaHTTPPort),
			},
		}, out...)
	}

	return out
}

func (wp *Wordpress) envFrom() []corev1.EnvFromSource {
	out := []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: wp.ComponentName(WordpressSecret),
				},
			},
		},
	}

	out = append(out, wp.Spec.EnvFrom...)

	return out
}

func (wp *Wordpress) gitCloneEnv() []corev1.EnvVar {
	if wp.Spec.CodeVolumeSpec.GitDir == nil {
		return []corev1.EnvVar{}
	}

	out := []corev1.EnvVar{
		{
			Name:  "GIT_CLONE_URL",
			Value: wp.Spec.CodeVolumeSpec.GitDir.Repository,
		},
		{
			Name:  "SRC_DIR",
			Value: codeSrcMountPath,
		},
	}

	if len(wp.Spec.CodeVolumeSpec.GitDir.GitRef) > 0 {
		out = append(out, corev1.EnvVar{
			Name:  "GIT_CLONE_REF",
			Value: wp.Spec.CodeVolumeSpec.GitDir.GitRef,
		})
	}

	out = append(out, wp.Spec.CodeVolumeSpec.GitDir.Env...)

	return out
}

func (wp *Wordpress) volumeMounts() (out []corev1.VolumeMount) {
	out = wp.Spec.VolumeMounts
	if wp.Spec.CodeVolumeSpec != nil {
		out = append(out, corev1.VolumeMount{
			Name:      codeVolumeName,
			MountPath: codeSrcMountPath,
			ReadOnly:  wp.Spec.CodeVolumeSpec.ReadOnly,
		})
		out = append(out, corev1.VolumeMount{
			Name:      codeVolumeName,
			MountPath: wp.Spec.CodeVolumeSpec.MountPath,
			ReadOnly:  wp.Spec.CodeVolumeSpec.ReadOnly,
			SubPath:   wp.Spec.CodeVolumeSpec.ContentSubPath,
		})
	}
	return out
}

func (wp *Wordpress) codeVolume() corev1.Volume {
	codeVolume := corev1.Volume{
		Name: codeVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}

	if wp.Spec.CodeVolumeSpec != nil {
		switch {
		case wp.Spec.CodeVolumeSpec.GitDir != nil:
			if wp.Spec.CodeVolumeSpec.GitDir.EmptyDir != nil {
				codeVolume.EmptyDir = wp.Spec.CodeVolumeSpec.GitDir.EmptyDir
			}
		case wp.Spec.CodeVolumeSpec.PersistentVolumeClaim != nil:
			codeVolume = corev1.Volume{
				Name: codeVolumeName,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: wp.ComponentName(WordpressCodePVC),
					},
				},
			}
		case wp.Spec.CodeVolumeSpec.HostPath != nil:
			codeVolume = corev1.Volume{
				Name: codeVolumeName,
				VolumeSource: corev1.VolumeSource{
					HostPath: wp.Spec.CodeVolumeSpec.HostPath,
				},
			}
		case wp.Spec.CodeVolumeSpec.EmptyDir != nil:
			codeVolume.EmptyDir = wp.Spec.CodeVolumeSpec.EmptyDir
		}
	}

	return codeVolume
}

func (wp *Wordpress) mediaVolume() corev1.Volume {
	mediaVolume := corev1.Volume{
		Name: mediaVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}

	if wp.Spec.MediaVolumeSpec != nil {
		switch {
		case wp.Spec.MediaVolumeSpec.PersistentVolumeClaim != nil:
			mediaVolume = corev1.Volume{
				Name: mediaVolumeName,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: wp.ComponentName(WordpressMediaPVC),
					},
				},
			}
		case wp.Spec.MediaVolumeSpec.HostPath != nil:
			mediaVolume = corev1.Volume{
				Name: mediaVolumeName,
				VolumeSource: corev1.VolumeSource{
					HostPath: wp.Spec.MediaVolumeSpec.HostPath,
				},
			}
		case wp.Spec.MediaVolumeSpec.EmptyDir != nil:
			mediaVolume.EmptyDir = wp.Spec.MediaVolumeSpec.EmptyDir
		}
	}

	return mediaVolume
}

func (wp *Wordpress) volumes() []corev1.Volume {
	return append(wp.Spec.Volumes, wp.codeVolume(), wp.mediaVolume())
}

func (wp *Wordpress) gitCloneContainer() corev1.Container {
	return corev1.Container{
		Name:    "git",
		Args:    []string{"/bin/bash", "-c", gitCloneScript},
		Image:   gitCloneImage,
		Env:     wp.gitCloneEnv(),
		EnvFrom: wp.Spec.CodeVolumeSpec.GitDir.EnvFrom,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      codeVolumeName,
				MountPath: codeSrcMountPath,
			},
		},
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &wwwDataUserID,
		},
	}
}

func (wp *Wordpress) rcloneContainer(name string, args []string) corev1.Container {
	var env []corev1.EnvVar
	var stream string

	switch {
	case wp.Spec.MediaVolumeSpec.S3VolumeSource != nil:
		env = wp.Spec.MediaVolumeSpec.S3VolumeSource.Env
		bucket := fmt.Sprintf("%s:%s", s3Prefix, wp.Spec.MediaVolumeSpec.S3VolumeSource.Bucket)
		stream = path.Join(bucket, wp.Spec.MediaVolumeSpec.S3VolumeSource.PathPrefix)

	case wp.Spec.MediaVolumeSpec.GCSVolumeSource != nil:
		env = wp.Spec.MediaVolumeSpec.GCSVolumeSource.Env
		bucket := fmt.Sprintf("%s:%s", gcsPrefix, wp.Spec.MediaVolumeSpec.GCSVolumeSource.Bucket)
		stream = path.Join(bucket, wp.Spec.MediaVolumeSpec.GCSVolumeSource.PathPrefix)
	}

	env = append(env, corev1.EnvVar{
		Name:  "RCLONE_STREAM",
		Value: stream,
	})

	return corev1.Container{
		Name:  name,
		Image: rcloneImage,
		Args:  args,
		Env:   env,
	}
}

func (wp *Wordpress) mediaContainers() []corev1.Container {
	if !wp.hasExternalMedia() {
		return []corev1.Container{}
	}

	// rclone-ftp
	// rclone serve ftp --vfs-cache-max-age 30s --vfs-cache-mde full --vfs-cache-poll-interval 0 --poll-interval 0
	// We want to cache writes and reads on the FTP server since thumbnails are going to be generated and also
	// because Wordpress is doing a directory listing in order to display the media gallery.
	// We also set the poll interval to zero to avoid any unnecessary requests to the remote buckets.
	ftpCmd := []string{
		"serve", "ftp", "-vvv", "--vfs-cache-max-age", "30s", "--vfs-cache-mode", "full",
		"--vfs-cache-poll-interval", "0", "--poll-interval", "0", "$(RCLONE_STREAM)/",
		fmt.Sprintf("--addr=0.0.0.0:%d", mediaFTPPort),
	}

	// rclone-http
	// rclone serve http --dir-cache-time 0
	// HTTP is used only to read media and the caching will be done in another layer. Also, because some upstreams are
	// eventually consistent, we don't want to cache stale responses.
	httpCmd := []string{
		"serve", "http", "-vvv", "--dir-cache-time", "0", "$(RCLONE_STREAM)/",
		fmt.Sprintf("--addr=0.0.0.0:%d", mediaHTTPPort),
	}

	return []corev1.Container{
		wp.rcloneContainer("rclone-ftp", ftpCmd),
		wp.rcloneContainer("rclone-http", httpCmd),
	}
}

func (wp *Wordpress) initContainers() []corev1.Container {
	containers := []corev1.Container{}

	if wp.hasExternalMedia() {
		// rclone-init-ftp
		// rclone touch gcs:prefix/wp-content/uploads/.keep
		// Because of https://bugs.php.net/bug.php?id=77680, we need to create the root directories.
		// For now, we don't support custom UPLOADS paths, only the default one (wp-content/uploads).
		// TODO: remove it once the fix is released
		initFTPCmd := []string{
			"touch", "-vvv", "$(RCLONE_STREAM)/wp-content/uploads/.keep",
		}

		containers = append(containers, wp.rcloneContainer("rclone-init-ftp", initFTPCmd))
	}

	if wp.Spec.CodeVolumeSpec != nil && wp.Spec.CodeVolumeSpec.GitDir != nil {
		containers = append(containers, wp.gitCloneContainer())
	}

	return containers
}

// WebPodTemplateSpec generates a pod template spec suitable for use in Wordpress deployment
func (wp *Wordpress) WebPodTemplateSpec() (out corev1.PodTemplateSpec) {
	out = corev1.PodTemplateSpec{}
	out.ObjectMeta.Labels = wp.WebPodLabels()

	out.Spec.ImagePullSecrets = wp.Spec.ImagePullSecrets
	if len(wp.Spec.ServiceAccountName) > 0 {
		out.Spec.ServiceAccountName = wp.Spec.ServiceAccountName
	}

	out.Spec.InitContainers = wp.initContainers()
	out.Spec.Containers = []corev1.Container{
		{
			Name:         "wordpress",
			Image:        wp.image(),
			VolumeMounts: wp.volumeMounts(),
			Env:          wp.env(),
			EnvFrom:      wp.envFrom(),
			Resources:    wp.Spec.Resources,
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: int32(InternalHTTPPort),
				},
			},
		},
	}
	out.Spec.Containers = append(out.Spec.Containers, wp.mediaContainers()...)

	out.Spec.Volumes = wp.volumes()

	if len(wp.Spec.NodeSelector) > 0 {
		out.Spec.NodeSelector = wp.Spec.NodeSelector
	}

	if len(wp.Spec.Tolerations) > 0 {
		out.Spec.Tolerations = wp.Spec.Tolerations
	}

	out.Spec.Affinity = wp.Spec.Affinity

	if len(wp.Spec.PriorityClassName) > 0 {
		out.Spec.PriorityClassName = wp.Spec.PriorityClassName
	}

	out.Spec.SecurityContext = &corev1.PodSecurityContext{
		FSGroup: &wwwDataUserID,
	}

	return out
}

// JobPodTemplateSpec generates a pod template spec suitable for use in wp-cli jobs
func (wp *Wordpress) JobPodTemplateSpec(cmd ...string) (out corev1.PodTemplateSpec) {
	out = corev1.PodTemplateSpec{}
	out.ObjectMeta.Labels = wp.JobPodLabels()

	out.Spec.ImagePullSecrets = wp.Spec.ImagePullSecrets
	if len(wp.Spec.ServiceAccountName) > 0 {
		out.Spec.ServiceAccountName = wp.Spec.ServiceAccountName
	}

	out.Spec.RestartPolicy = corev1.RestartPolicyNever

	out.Spec.InitContainers = wp.initContainers()
	out.Spec.Containers = []corev1.Container{
		{
			Name:         "wp-cli",
			Image:        wp.image(),
			Args:         cmd,
			VolumeMounts: wp.volumeMounts(),
			Env:          wp.env(),
			EnvFrom:      wp.envFrom(),
		},
	}
	out.Spec.Containers = append(out.Spec.Containers, wp.mediaContainers()...)

	out.Spec.Volumes = wp.volumes()

	if len(wp.Spec.NodeSelector) > 0 {
		out.Spec.NodeSelector = wp.Spec.NodeSelector
	}

	if len(wp.Spec.Tolerations) > 0 {
		out.Spec.Tolerations = wp.Spec.Tolerations
	}

	out.Spec.Affinity = wp.Spec.Affinity

	if len(wp.Spec.PriorityClassName) > 0 {
		out.Spec.PriorityClassName = wp.Spec.PriorityClassName
	}

	out.Spec.SecurityContext = &corev1.PodSecurityContext{
		FSGroup: &wwwDataUserID,
	}

	return out
}
