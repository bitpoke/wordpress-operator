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
	"hash/fnv"
	"strings"

	core_util "github.com/appscode/kutil/core/v1"
	corev1 "k8s.io/api/core/v1"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
)

const (
	webrootVolumeName = "webroot"
	mediaVolumeName   = "media"
)

var (
	defaultImage = "defaultImage"
)

// GetImage return the image for the Wordpress resource relative to the WordpressRuntime
func GetImage(wp *wordpressv1alpha1.Wordpress, rt *wordpressv1alpha1.WordpressRuntime) string {
	image := rt.Spec.DefaultImage
	if len(wp.Spec.Image) > 0 {
		image = wp.Spec.Image
	}
	return image
}

// GetVersionHash returns the Wordpress image version hash which can be used in kubernetes resource names
func GetVersionHash(wp *wordpressv1alpha1.Wordpress, rt *wordpressv1alpha1.WordpressRuntime) string {
	image := GetImage(wp, rt)
	return fmt.Sprintf("%x", fnv.New32a().Sum([]byte(image)))[:32]
}

// WebPodTemplateSpec generates a pod template spec suitable for use in Wordpress deployment
func WebPodTemplateSpec(wp *wordpressv1alpha1.Wordpress, rt *wordpressv1alpha1.WordpressRuntime) (out corev1.PodTemplateSpec) {
	if rt.Spec.WebPodTemplate != nil {
		rt.Spec.WebPodTemplate.DeepCopyInto(&out)
	} else {
		out = corev1.PodTemplateSpec{}
	}

	out.ObjectMeta.Labels = WebPodLabels(wp)

	out.Spec.Volumes = ensureWordpressVolumes(wp, rt, out.Spec.Volumes)

	for i := range out.Spec.InitContainers {
		ensureWordpressEnv(wp, &out.Spec.InitContainers[i])
		ensureWordpressVolumeMounts(wp, &out.Spec.InitContainers[i])
		setContainerImage(wp, rt, &out.Spec.InitContainers[i])
	}
	for i := range out.Spec.Containers {
		ensureWordpressEnv(wp, &out.Spec.Containers[i])
		ensureWordpressVolumeMounts(wp, &out.Spec.Containers[i])
		setContainerImage(wp, rt, &out.Spec.Containers[i])
	}

	out.Spec.ImagePullSecrets = append(out.Spec.ImagePullSecrets, wp.Spec.ImagePullSecrets...)
	if len(wp.Spec.ServiceAccountName) > 0 {
		out.Spec.ServiceAccountName = wp.Spec.ServiceAccountName
	}
	return out
}

// JobPodTemplateSpec generates a pod template spec suitable for WP CLI background jobs
func JobPodTemplateSpec(wp *wordpressv1alpha1.Wordpress, rt *wordpressv1alpha1.WordpressRuntime, cmd ...string) (out corev1.PodTemplateSpec) {
	if rt.Spec.CLIPodTemplate != nil {
		rt.Spec.CLIPodTemplate.DeepCopyInto(&out)
	} else {
		out = corev1.PodTemplateSpec{}
	}

	out.ObjectMeta.Labels = JobPodLabels(wp)

	out.Spec.Volumes = ensureWordpressVolumes(wp, rt, out.Spec.Volumes)
	out.Spec.RestartPolicy = corev1.RestartPolicyNever

	for i := range out.Spec.InitContainers {
		ensureWordpressEnv(wp, &out.Spec.InitContainers[i])
		ensureWordpressVolumeMounts(wp, &out.Spec.InitContainers[i])
		setContainerImage(wp, rt, &out.Spec.InitContainers[i])
	}
	for i := range out.Spec.Containers {
		ensureWordpressEnv(wp, &out.Spec.Containers[i])
		ensureWordpressVolumeMounts(wp, &out.Spec.Containers[i])
		setContainerImage(wp, rt, &out.Spec.Containers[i])
	}

	for i, c := range out.Spec.Containers {
		if c.Name == "wp-cli" {
			out.Spec.Containers[i].Args = append(out.Spec.Containers[i].Args, cmd...)
		}
	}

	out.Spec.ImagePullSecrets = append(out.Spec.ImagePullSecrets, wp.Spec.ImagePullSecrets...)
	if len(wp.Spec.ServiceAccountName) > 0 {
		out.Spec.ServiceAccountName = wp.Spec.ServiceAccountName
	}
	return out
}

func ensureWordpressEnv(wp *wordpressv1alpha1.Wordpress, ctr *corev1.Container) {
	ctr.Env = core_util.UpsertEnvVars(ctr.Env, wp.Spec.Env...)

	domains := make([]string, len(wp.Spec.Domains))
	for i, d := range wp.Spec.Domains {
		domains[i] = string(d)
	}
	env := []corev1.EnvVar{
		{
			Name:  "WORDPRESS_DOMAINS",
			Value: strings.Join(domains, ","),
		},
	}

	ctr.Env = core_util.UpsertEnvVars(ctr.Env, env...)
	ctr.EnvFrom = append(wp.Spec.EnvFrom, ctr.EnvFrom...)
}

func ensureWordpressVolumeMounts(wp *wordpressv1alpha1.Wordpress, ctr *corev1.Container) {
	for _, v := range wp.Spec.VolumeMountsSpec {
		ctr.VolumeMounts = core_util.UpsertVolumeMountByPath(ctr.VolumeMounts, v)
	}
}

func ensureVolume(name, pvcName string, volSpec *wordpressv1alpha1.WordpressVolumeSpec, in []corev1.Volume) []corev1.Volume {
	if volSpec == nil {
		return in
	}
	v := corev1.Volume{Name: name}

	if volSpec.PersistentVolumeClaim != nil {
		v.VolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvcName,
			},
		}
	} else if volSpec.HostPath != nil {
		v.VolumeSource = corev1.VolumeSource{HostPath: volSpec.HostPath}
	} else {
		d := volSpec.EmptyDir
		if d == nil {
			d = &corev1.EmptyDirVolumeSource{}
		}
		v.VolumeSource = corev1.VolumeSource{EmptyDir: d}
	}

	in = core_util.UpsertVolume(in, v)
	return in
}

func ensureWordpressVolumes(wp *wordpressv1alpha1.Wordpress, rt *wordpressv1alpha1.WordpressRuntime, in []corev1.Volume) []corev1.Volume {
	for _, vol := range wp.Spec.Volumes {
		in = core_util.UpsertVolume(in, vol)
	}

	// webroot (plugins, themes, etc.)
	volSpec := rt.Spec.WebrootVolumeSpec
	if wp.Spec.WebrootVolumeSpec != nil {
		volSpec = wp.Spec.WebrootVolumeSpec
	}
	in = ensureVolume(webrootVolumeName, fmt.Sprintf("%s-webroot", wp.Name), volSpec, in)

	volSpec = rt.Spec.MediaVolumeSpec
	if wp.Spec.MediaVolumeSpec != nil {
		volSpec = wp.Spec.MediaVolumeSpec
	}
	in = ensureVolume(mediaVolumeName, fmt.Sprintf("%s-webroot", wp.Name), volSpec, in)

	return in
}

func setContainerImage(wp *wordpressv1alpha1.Wordpress, rt *wordpressv1alpha1.WordpressRuntime, ctr *corev1.Container) {
	if ctr.Image != defaultImage {
		return
	}

	image := GetImage(wp, rt)
	imagePullPolicy := rt.Spec.DefaultImagePullPolicy
	if len(wp.Spec.ImagePullPolicy) > 0 {
		imagePullPolicy = wp.Spec.ImagePullPolicy
	}

	ctr.Image = image
	if len(imagePullPolicy) > 0 {
		ctr.ImagePullPolicy = imagePullPolicy
	}
}
