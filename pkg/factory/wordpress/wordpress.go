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
	"fmt"
	"strings"

	core_util "github.com/appscode/kutil/core/v1"
	// "github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	wpapi "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
)

const (
	cronName       = "%s-wp-cron"
	deploymentName = "%s"
	serviceName    = "%s"
	ingressName    = "%s"
	contentPVCName = "%s"
	mediaPVCName   = "%s-media"

	contentVolumeName = "content"
	mediaVolumeName   = "media"
)

var contentMountPaths = []string{"root", "themes", "plugins", "mu-plugins", "languages", "upgrade"}
var tempMountPaths = []string{"run", "tmp"}

type Generator struct {
	WP *wpapi.Wordpress
}

// Labels returns a general label set to apply to objects, relative to the
// Wordpress API object
func (g *Generator) Labels() labels.Set {
	return labels.Set{
		"app.kubernetes.io/name":           "wordpress",
		"app.kubernetes.io/app-instance":   g.WP.ObjectMeta.Name,
		"app.kubernetes.io/deploy-manager": "wordpress-operator",
	}
}

func (g *Generator) WebPodLabels() labels.Set {
	l := g.Labels()
	l["app.kubernetes.io/component"] = "web"
	l["app.kubernetes.io/tier"] = "front"

	return l
}

func (g *Generator) DeploymentName() string {
	return fmt.Sprintf(deploymentName, g.WP.Name)
}

func (g *Generator) ContentPVCName() string {
	return fmt.Sprintf(contentPVCName, g.WP.Name)
}

func (g *Generator) MediaPVCName() string {
	return fmt.Sprintf(mediaPVCName, g.WP.Name)
}

func (g *Generator) ServiceName() string {
	return fmt.Sprintf(serviceName, g.WP.Name)
}

func (g *Generator) IngressName() string {
	return fmt.Sprintf(ingressName, g.WP.Name)
}

func (g *Generator) WPCronName() string {
	return fmt.Sprintf(cronName, g.WP.Name)
}

// PodTemplateSpec generates a pod template spec suitable for use in Wordpress
// deployment
func (g *Generator) WebPodTemplateSpec(in *corev1.PodTemplateSpec) (out *corev1.PodTemplateSpec) {
	if g.WP.Spec.WebPodTemplate != nil {
		out = g.WP.Spec.WebPodTemplate.DeepCopy()
	} else {
		out = &corev1.PodTemplateSpec{}
	}

	out.ObjectMeta.Labels = g.WebPodLabels()

	out.Spec.Volumes = g.ensureWordpressVolumes(out.Spec.Volumes)

	for i := range out.Spec.InitContainers {
		g.ensureWordpressEnv(&out.Spec.InitContainers[i])
		g.ensureWordpressVolumeMounts(&out.Spec.InitContainers[i])
	}
	for i := range out.Spec.Containers {
		g.ensureWordpressEnv(&out.Spec.Containers[i])
		g.ensureWordpressVolumeMounts(&out.Spec.Containers[i])
	}

	return
}

// JobPodTemplate generates a pod template spec suitable for WP CLI background
// jobs
func (g *Generator) JobPodTemplateSpec(in *corev1.PodTemplateSpec, cmd ...string) (out *corev1.PodTemplateSpec) {
	if g.WP.Spec.WebPodTemplate != nil {
		out = g.WP.Spec.CLIPodTemplate.DeepCopy()
	} else {
		out = &corev1.PodTemplateSpec{}
	}

	out.ObjectMeta.Labels = g.WebPodLabels()

	out.Spec.Volumes = g.ensureWordpressVolumes(out.Spec.Volumes)

	for i := range out.Spec.InitContainers {
		g.ensureWordpressEnv(&out.Spec.InitContainers[i])
		g.ensureWordpressVolumeMounts(&out.Spec.InitContainers[i])
	}
	for i := range out.Spec.Containers {
		g.ensureWordpressEnv(&out.Spec.Containers[i])
		g.ensureWordpressVolumeMounts(&out.Spec.Containers[i])
	}

	return
}

func (g *Generator) ensureWordpressEnv(ctr *corev1.Container) {
	ctr.Env = core_util.UpsertEnvVars(ctr.Env, g.WP.Spec.Env...)

	var domains []string
	for _, d := range g.WP.Spec.Domains {
		domains = append(domains, string(d))
	}
	env := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "WORDPRESS_DOMAINS",
			Value: strings.Join(domains, ","),
		},
	}

	ctr.Env = core_util.UpsertEnvVars(ctr.Env, env...)
	ctr.EnvFrom = append(g.WP.Spec.EnvFrom, ctr.EnvFrom...)
}

func (g *Generator) ensureWordpressVolumeMounts(ctr *corev1.Container) {
	for _, v := range g.WP.Spec.VolumeMountsSpec {
		ctr.VolumeMounts = core_util.UpsertVolumeMountByPath(ctr.VolumeMounts, v)
	}
}

func (g *Generator) ensureWordpressVolumes(in []corev1.Volume) []corev1.Volume {
	var v corev1.Volume

	// content (plugins, themes, etc.)
	v = corev1.Volume{Name: contentVolumeName}
	if g.WP.Spec.ContentVolumeSpec.PersistentVolumeClaim != nil {
		v.VolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: fmt.Sprintf(contentPVCName, g.WP.Name),
			},
		}
	} else if g.WP.Spec.ContentVolumeSpec.HostPath != nil {
		v.VolumeSource = corev1.VolumeSource{HostPath: g.WP.Spec.ContentVolumeSpec.HostPath}
	} else {
		d := g.WP.Spec.ContentVolumeSpec.EmptyDir
		if d == nil {
			d = &corev1.EmptyDirVolumeSource{}
		}
		v.VolumeSource = corev1.VolumeSource{EmptyDir: d}
	}
	in = core_util.UpsertVolume(in, v)

	// media files
	if g.WP.Spec.MediaVolumeSpec == nil {
		in = core_util.EnsureVolumeDeleted(in, mediaVolumeName)
		// Return is MediaVolumeSpec is not defined
		return in
	}

	v = corev1.Volume{Name: mediaVolumeName}
	if g.WP.Spec.MediaVolumeSpec.PersistentVolumeClaim != nil {
		v.VolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: fmt.Sprintf(contentPVCName, g.WP.Name),
			},
		}
	} else if g.WP.Spec.MediaVolumeSpec.HostPath != nil {
		v.VolumeSource = corev1.VolumeSource{HostPath: g.WP.Spec.MediaVolumeSpec.HostPath}
	} else {
		d := g.WP.Spec.MediaVolumeSpec.EmptyDir
		if d == nil {
			d = &corev1.EmptyDirVolumeSource{}
		}
		v.VolumeSource = corev1.VolumeSource{EmptyDir: d}
	}
	in = core_util.UpsertVolume(in, v)
	return in
}

func (g *Generator) ensureContainerDefaults(ctr *corev1.Container) {
	if len(ctr.Resources.Limits) == 0 {
		ctr.Resources.Limits = make(corev1.ResourceList)
	}
	if len(ctr.Resources.Requests) == 0 {
		ctr.Resources.Requests = make(corev1.ResourceList)
	}
}
