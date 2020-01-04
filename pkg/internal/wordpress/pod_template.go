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

	"path"

	"strings"

	"github.com/presslabs/wordpress-operator/pkg/cmd/options"
)

const (
	// InternalHTTPPort represents the internal port used by the runtime container
	InternalHTTPPort = 8080
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
git checkout -B "$GIT_CLONE_REF" "origin/$GIT_CLONE_REF"
`

var (
	s3EnvVars = map[string]string{
		"AWS_ACCESS_KEY_ID":     "AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY": "AWS_SECRET_ACCESS_KEY",
		"AWS_CONFIG_FILE":       "AWS_CONFIG_FILE",
		"ENDPOINT":              "S3_ENDPOINT",
	}
	gcsEnvVars = map[string]string{
		"GOOGLE_CREDENTIALS":             "GOOGLE_CREDENTIALS",
		"GOOGLE_APPLICATION_CREDENTIALS": "GOOGLE_APPLICATION_CREDENTIALS",
	}
)

func (wp *Wordpress) mediaEnv() []corev1.EnvVar {
	out := []corev1.EnvVar{}

	if wp.Spec.MediaVolumeSpec == nil {
		return out
	}

	if wp.Spec.MediaVolumeSpec.S3VolumeSource != nil {
		bucket := path.Join(wp.Spec.MediaVolumeSpec.S3VolumeSource.Bucket, wp.Spec.MediaVolumeSpec.S3VolumeSource.PathPrefix)
		out = append(out, corev1.EnvVar{
			Name:  "STACK_MEDIA_BUCKET",
			Value: fmt.Sprintf("%s://%s", s3Prefix, bucket),
		})
		for _, env := range wp.Spec.MediaVolumeSpec.S3VolumeSource.Env {
			if name, ok := s3EnvVars[env.Name]; ok {
				_env := env.DeepCopy()
				_env.Name = name
				out = append(out, *_env)
			}
		}
	}

	if wp.Spec.MediaVolumeSpec.GCSVolumeSource != nil {
		bucket := path.Join(wp.Spec.MediaVolumeSpec.GCSVolumeSource.Bucket, wp.Spec.MediaVolumeSpec.GCSVolumeSource.PathPrefix)
		out = append(out, corev1.EnvVar{
			Name:  "STACK_MEDIA_BUCKET",
			Value: fmt.Sprintf("%s://%s", gcsPrefix, bucket),
		})
		for _, env := range wp.Spec.MediaVolumeSpec.GCSVolumeSource.Env {
			if name, ok := gcsEnvVars[env.Name]; ok {
				_env := env.DeepCopy()
				_env.Name = name
				out = append(out, *_env)
			}
		}
	}

	return out

}

func (wp *Wordpress) routes() []string {
	if len(wp.Spec.Routes) == 0 {
		return []string{wp.MainDomain()}
	}
	var out = make([]string, len(wp.Spec.Routes))
	for i, r := range wp.Spec.Routes {
		out[i] = path.Join(r.Domain, r.Path)
	}
	return out
}

func (wp *Wordpress) env() []corev1.EnvVar {
	out := append([]corev1.EnvVar{
		{
			Name:  "WP_HOME",
			Value: wp.HomeURL(),
		},
		{
			Name:  "WP_SITEURL",
			Value: wp.HomeURL("wp"),
		},
		{
			Name:  "STACK_ROUTES",
			Value: strings.Join(wp.routes(), ","),
		},
	}, wp.Spec.Env...)

	out = append(out, wp.mediaEnv()...)

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
	if wp.hasCodeMounts() {
		out = append(out, corev1.VolumeMount{
			MountPath: codeSrcMountPath,
			Name:      codeVolumeName,
			ReadOnly:  wp.Spec.CodeVolumeSpec.ReadOnly,
		})
		out = append(out, corev1.VolumeMount{
			MountPath: wp.Spec.CodeVolumeSpec.MountPath,
			Name:      codeVolumeName,
			ReadOnly:  wp.Spec.CodeVolumeSpec.ReadOnly,
			SubPath:   wp.Spec.CodeVolumeSpec.ContentSubPath,
		})
		out = append(out, corev1.VolumeMount{
			MountPath: configMountPath,
			Name:      codeVolumeName,
			ReadOnly:  true,
			SubPath:   wp.Spec.CodeVolumeSpec.ConfigSubPath,
		})
	}
	if wp.hasMediaMounts() {
		v := corev1.VolumeMount{
			MountPath: wp.Spec.MediaVolumeSpec.MountPath,
			Name:      mediaVolumeName,
			ReadOnly:  wp.Spec.MediaVolumeSpec.ReadOnly,
		}
		if wp.Spec.MediaVolumeSpec.ContentSubPath != "" {
			v.SubPath = wp.Spec.MediaVolumeSpec.ContentSubPath
		}
		out = append(out, v)
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
	volumes := wp.Spec.Volumes
	if wp.hasCodeMounts() {
		volumes = append(volumes, wp.codeVolume())
	}
	if wp.hasMediaMounts() {
		volumes = append(volumes, wp.mediaVolume())
	}
	return volumes
}

func (wp *Wordpress) securityContext() *corev1.SecurityContext {
	defaultProcMount := corev1.DefaultProcMount
	return &corev1.SecurityContext{
		RunAsUser: wp.Spec.RunAsUser,
		ProcMount: &defaultProcMount,
	}
}

func (wp *Wordpress) gitCloneContainer() corev1.Container {
	return corev1.Container{
		Name:    "git",
		Args:    []string{"/bin/bash", "-c", gitCloneScript},
		Image:   options.GitCloneImage,
		Env:     wp.gitCloneEnv(),
		EnvFrom: wp.Spec.CodeVolumeSpec.GitDir.EnvFrom,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      codeVolumeName,
				MountPath: codeSrcMountPath,
			},
		},
		SecurityContext: wp.securityContext(),
	}
}

func (wp *Wordpress) installWPContainer() []corev1.Container {
	if wp.Spec.WordpressBootstrapSpec == nil {
		return []corev1.Container{}
	}

	return []corev1.Container{
		{
			Name:            "install-wp",
			Image:           wp.Spec.Image,
			VolumeMounts:    wp.volumeMounts(),
			Env:             append(wp.env(), wp.Spec.WordpressBootstrapSpec.Env...),
			EnvFrom:         append(wp.envFrom(), wp.Spec.WordpressBootstrapSpec.EnvFrom...),
			SecurityContext: wp.securityContext(),
			Command:         []string{"wp-install"},
			Args: []string{
				"$(WORDPRESS_BOOTSTRAP_TITLE)",
				wp.HomeURL(),
				"$(WORDPRESS_BOOTSTRAP_USER)",
				"$(WORDPRESS_BOOTSTRAP_PASSWORD)",
				"$(WORDPRESS_BOOTSTRAP_EMAIL)",
			},
		},
	}
}

func (wp *Wordpress) initContainers() []corev1.Container {
	containers := []corev1.Container{}

	if wp.Spec.CodeVolumeSpec != nil && wp.Spec.CodeVolumeSpec.GitDir != nil {
		containers = append(containers, wp.gitCloneContainer())
	}

	// first clone data then install wp
	containers = append(containers, wp.installWPContainer()...)

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
			Name:            "wordpress",
			Image:           wp.Spec.Image,
			ImagePullPolicy: wp.Spec.ImagePullPolicy,
			VolumeMounts:    wp.volumeMounts(),
			Env:             wp.env(),
			EnvFrom:         wp.envFrom(),
			Resources:       wp.Spec.Resources,
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: int32(InternalHTTPPort),
				},
			},
			SecurityContext: wp.securityContext(),
		},
	}

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
		FSGroup: wp.Spec.RunAsUser,
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
			Name:            "wp-cli",
			Image:           wp.Spec.Image,
			ImagePullPolicy: wp.Spec.ImagePullPolicy,
			Args:            cmd,
			VolumeMounts:    wp.volumeMounts(),
			Env:             wp.env(),
			EnvFrom:         wp.envFrom(),
			SecurityContext: wp.securityContext(),
		},
	}

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
		FSGroup: wp.Spec.RunAsUser,
	}

	return out
}

func (wp *Wordpress) hasMediaMounts() bool {
	if wp.Spec.MediaVolumeSpec == nil {
		return false
	}
	switch {
	case wp.Spec.MediaVolumeSpec.PersistentVolumeClaim != nil:
		return true
	case wp.Spec.MediaVolumeSpec.HostPath != nil:
		return true
	case wp.Spec.MediaVolumeSpec.EmptyDir != nil:
		return true
	}
	return false
}

func (wp *Wordpress) hasCodeMounts() bool {
	if wp.Spec.CodeVolumeSpec == nil {
		return false
	}
	switch {
	case wp.Spec.CodeVolumeSpec.GitDir != nil:
		return true
	case wp.Spec.CodeVolumeSpec.PersistentVolumeClaim != nil:
		return true
	case wp.Spec.CodeVolumeSpec.HostPath != nil:
		return true
	case wp.Spec.CodeVolumeSpec.EmptyDir != nil:
		return true
	}
	return false
}
