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
	gitCloneImage     = "quay.io/presslabs/git-clone:latest"
	wordpressHTTPPort = 80
	codeVolumeName    = "code"
	mediaVolumeName   = "media"
)

var (
	s3EnvVars = map[string]string{
		"ACCESS_KEY":        "WORDPRESS_S3_ACCESS_KEY",
		"SECRET_ACCESS_KEY": "WORDPRESS_S3_SECRET_KEY",
		"ENDPOINT":          "WORDPRESS_S3_ENDPOINT",
	}
	gcsEnvVars = map[string]string{
		"GOOGLE_APPLICATION_CREDENTIALS_JSON": "WORDPRESS_GCS_APPLICATION_CREDENTIALS_JSON",
	}
)

func (wp *Wordpress) image() string {
	return fmt.Sprintf("%s:%s", wp.Spec.Image, wp.Spec.Tag)
}

func (wp *Wordpress) env() []corev1.EnvVar {
	out := wp.Spec.Env
	if wp.Spec.MediaVolumeSpec != nil {
		if wp.Spec.MediaVolumeSpec.S3VolumeSource != nil {
			for _, env := range wp.Spec.MediaVolumeSpec.S3VolumeSource.Env {
				if name, ok := s3EnvVars[env.Name]; ok {
					_env := env.DeepCopy()
					_env.Name = name
					wp.Spec.Env = append(wp.Spec.Env, *_env)
				}
			}
		}

		if wp.Spec.MediaVolumeSpec.GCSVolumeSource != nil {
			for _, env := range wp.Spec.MediaVolumeSpec.GCSVolumeSource.Env {
				if name, ok := gcsEnvVars[env.Name]; ok {
					_env := env.DeepCopy()
					_env.Name = name
					wp.Spec.Env = append(wp.Spec.Env, *_env)
				}
			}
		}
	}
	return out
}

func (wp *Wordpress) volumeMounts() (out []corev1.VolumeMount) {
	out = wp.Spec.VolumeMounts
	if wp.Spec.CodeVolumeSpec != nil {
		out = append(out, corev1.VolumeMount{
			Name:      codeVolumeName,
			MountPath: "/code",
			ReadOnly:  wp.Spec.CodeVolumeSpec.ReadOnly,
		})
		out = append(out, corev1.VolumeMount{
			Name:      codeVolumeName,
			MountPath: "/var/www/html/wp-content",
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
			{
				Name:  "git",
				Image: gitCloneImage,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      codeVolumeName,
						MountPath: "/code",
					},
				},
			},
		}
	}

	out.Spec.Containers = []corev1.Container{
		{
			Name:         "wordpress",
			Image:        wp.image(),
			VolumeMounts: wp.volumeMounts(),
			Env:          wp.env(),
			EnvFrom:      wp.Spec.EnvFrom,
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: int32(wordpressHTTPPort),
				},
			},
		},
	}

	out.Spec.Volumes = wp.volumes()

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
			{
				Name:  "git",
				Image: gitCloneImage,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      codeVolumeName,
						MountPath: "/code",
					},
				},
			},
		}
	}

	out.Spec.Containers = []corev1.Container{
		{
			Name:         "wp-cli",
			Image:        wp.image(),
			Args:         cmd,
			VolumeMounts: wp.volumeMounts(),
			Env:          wp.env(),
			EnvFrom:      wp.Spec.EnvFrom,
		},
	}

	out.Spec.Volumes = wp.volumes()

	return out
}
