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

	corev1 "k8s.io/api/core/v1"
)

const (
	gitCloneImage     = "docker.io/library/buildpack-deps:stretch-scm"
	wordpressHTTPPort = 80
	codeVolumeName    = "code"
	mediaVolumeName   = "media"
)

const rcloneImage = "quay.io/presslabs/rclone"

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

var (
	s3EnvVars = map[string]string{
		"AWS_ACCESS_KEY_ID":     "AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY": "AWS_SECRET_ACCESS_KEY",
		"AWS_CONFIG_FILE":       "AWS_CONFIG_FILE",
		"ENDPOINT":              "S3_ENDPOINT",
		"REGION":              	 "AWS_REGION",
		"ACL":              	 "AWS_ACL",
		"STORAGE_CLASS":         "AWS_STORAGE_CLASS",
	}
	gcsEnvVars = map[string]string{
		"GOOGLE_CREDENTIALS":             "GCS_SERVICE_ACCOUNT_JSON_KEY",
		"GOOGLE_APPLICATION_CREDENTIALS": "GCS_SERVICE_ACCOUNT_JSON_KEY",
		"GOOGLE_PROJECT_ID":			  "GCS_PROJECT_ID",
		"GOOGLE_OBJECT_ACL": 			  "GCS_OBJECT_ACL",
		"GOOGLE_BUCKET_ACL": 			  "GCS_BUCKET_ACL",
		"GOOGLE_STORAGE_LOCATION": 		  "GCS_LOCATION",
		"GOOGLE_STORAGE_CLASS": 		  "GCS_STORAGE_CLASS",
	}
)

func (wp *Wordpress) image() string {
	return fmt.Sprintf("%s:%s", wp.Spec.Image, wp.Spec.Tag)
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

func (wp *Wordpress) getUploadsContainers() []corev1.Container {
	if wp.Spec.MediaVolumeSpec == nil ||
		(wp.Spec.MediaVolumeSpec.S3VolumeSource == nil && wp.Spec.MediaVolumeSpec.GCSVolumeSource == nil){
		return []corev1.Container{}
	}

	var stream string
	var env []corev1.EnvVar

	if wp.Spec.MediaVolumeSpec.S3VolumeSource != nil {
		stream = fmt.Sprintf("s3:%s/%s", wp.Spec.MediaVolumeSpec.S3VolumeSource.Bucket,
			wp.Spec.MediaVolumeSpec.S3VolumeSource.PathPrefix)

		for _, e := range wp.Spec.MediaVolumeSpec.S3VolumeSource.Env {
			if name, ok := s3EnvVars[e.Name]; ok {
				_env := e.DeepCopy()
				_env.Name = name
				env = append(env, *_env)
			}
		}
	} else if wp.Spec.MediaVolumeSpec.GCSVolumeSource != nil {
		stream = fmt.Sprintf("gs:%s/%s", wp.Spec.MediaVolumeSpec.GCSVolumeSource.Bucket,
			wp.Spec.MediaVolumeSpec.GCSVolumeSource.PathPrefix)

		for _, e := range wp.Spec.MediaVolumeSpec.GCSVolumeSource.Env {
			if name, ok := gcsEnvVars[e.Name]; ok {
				_env := e.DeepCopy()
				_env.Name = name
				env = append(env, *_env)
			}
		}
	}

	return []corev1.Container {
		{
			Name:    "rclone-ftp",
			Image:   rcloneImage,
			Args:    []string{"serve", "ftp", "--vfs-cache-max-age", "30s", "--vfs-cache-mode", "full",
				"--vfs-cache-poll-interval", "0", "--poll-interval", "0", stream, "--config=/etc/rclone.conf",
				"--addr=0.0.0.0:2121"},
			Env:     env,
		},
		{
			Name:    "rclone-http",
			Image:   rcloneImage,
			Args:    []string{"serve", "http", "--dir-cache-time", "0", stream, "--config=/etc/rclone.conf",
				"--addr=0.0.0.0:8080"},
			Env:     env,
		},
	}
}

// WebPodTemplateSpec generates a pod template spec suitable for use in Wordpress deployment
func (wp *Wordpress) WebPodTemplateSpec() (out corev1.PodTemplateSpec) {
	out = corev1.PodTemplateSpec{}
	out.ObjectMeta.Labels = wp.WebPodLabels()

	out.Spec.ImagePullSecrets = wp.Spec.ImagePullSecrets
	if len(wp.Spec.ServiceAccountName) > 0 {
		out.Spec.ServiceAccountName = wp.Spec.ServiceAccountName
	}

	if wp.Spec.CodeVolumeSpec != nil && wp.Spec.CodeVolumeSpec.GitDir != nil {
		out.Spec.InitContainers = []corev1.Container{
			wp.gitCloneContainer(),
		}
	}

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
					ContainerPort: int32(wordpressHTTPPort),
				},
			},
		},
	}
	out.Spec.Containers = append(out.Spec.Containers, wp.getUploadsContainers()...)

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

	if wp.Spec.CodeVolumeSpec != nil && wp.Spec.CodeVolumeSpec.GitDir != nil {
		out.Spec.InitContainers = []corev1.Container{
			wp.gitCloneContainer(),
		}
	}

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
	out.Spec.Containers = append(out.Spec.Containers, wp.getUploadsContainers()...)

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
