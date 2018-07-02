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

	"github.com/kballard/go-shellquote"

	core_util "github.com/appscode/kutil/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"

	wpapi "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
)

const (
	phpContainerName   = "wordpress"
	nginxContainerName = "nginx"
	contentVolumeName  = "content"
	mediaVolumeName    = "media"
	tempVolumeName     = "temp"
	tempVolumeSize     = 128 * 1024 * 1024
	contentPVCName     = "%s"
	mediaPVCName       = "%s-media"
)

var contentMountPaths = []string{"root", "themes", "plugins", "mu-plugins", "languages", "upgrade"}
var tempMountPaths = []string{"run", "tmp"}

type Generator struct {
	WP                  *wpapi.Wordpress
	DefaultRuntimeImage string
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

func (g *Generator) ContentVolumeName() string {
	return fmt.Sprintf(contentPVCName, g.WP.ObjectMeta.Name)
}

func (g *Generator) MediaVolumeName() string {
	return fmt.Sprintf(mediaPVCName, g.WP.ObjectMeta.Name)
}

// PodTemplateSpec generates a pod template spec suitable for use in Wordpress
// deployment
func (g *Generator) PodTemplateSpec(in *corev1.PodTemplateSpec) (out *corev1.PodTemplateSpec) {
	in.ObjectMeta.Labels = g.Labels()

	in.Spec.Containers = g.ensureNginxContainer(in.Spec.Containers)
	in.Spec.Containers = g.ensurePHPContainer(in.Spec.Containers)

	in.Spec.NodeSelector = g.WP.Spec.NodeSelector
	in.Spec.Tolerations = g.WP.Spec.Tolerations
	in.Spec.Affinity = &g.WP.Spec.Affinity

	if len(g.WP.Spec.ImagePullSecrets) > 0 {
		in.Spec.ImagePullSecrets = g.WP.Spec.ImagePullSecrets
	}
	if len(g.WP.Spec.ServiceAccountName) > 0 {
		in.Spec.ServiceAccountName = g.WP.Spec.ServiceAccountName
	} else {
		in.Spec.ServiceAccountName = "default"
	}

	in.Spec.Volumes = g.ensureVolumes(in.Spec.Volumes)

	return in
}

// JobPodTemplate generates a pod template spec suitable for use ing.WP CLI
func (g *Generator) JobPodTemplateSpec(in *corev1.PodTemplateSpec, cmd ...string) (out *corev1.PodTemplateSpec) {
	switch g.WP.Spec.CLIDriver {
	case "inline":
		in = g.inlineJobPodTemplate(in, cmd...)
	default:
		in = g.standaloneJobPodTemplate(in, cmd...)
	}

	in.Spec.RestartPolicy = "Never"

	if len(g.WP.Spec.ImagePullSecrets) > 0 {
		in.Spec.ImagePullSecrets = g.WP.Spec.ImagePullSecrets
	}

	if len(g.WP.Spec.ServiceAccountName) > 0 {
		in.Spec.ServiceAccountName = g.WP.Spec.ServiceAccountName
	} else {
		in.Spec.ServiceAccountName = "default"
	}

	return in
}

func (g *Generator) standaloneJobPodTemplate(in *corev1.PodTemplateSpec, cmd ...string) (out *corev1.PodTemplateSpec) {
	in.Spec.Containers = g.ensurePHPContainer(in.Spec.Containers)

	// cleanup entrypoint and enforce the requested command
	in.Spec.Containers[0].Command = []string{}
	in.Spec.Containers[0].Args = cmd

	// setup container resources
	in.Spec.Containers[0].Resources.Limits = make(corev1.ResourceList)
	in.Spec.Containers[0].Resources.Requests = make(corev1.ResourceList)
	if r, ok := g.WP.Spec.Resources.Limits[wpapi.ResourceCLICPU]; ok {
		in.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU] = r
	}
	if r, ok := g.WP.Spec.Resources.Limits[wpapi.ResourceCLIMemory]; ok {
		in.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] = r
	}
	if r, ok := g.WP.Spec.Resources.Requests[wpapi.ResourceCLICPU]; ok {
		in.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU] = r
	}
	if r, ok := g.WP.Spec.Resources.Requests[wpapi.ResourceCLIMemory]; ok {
		in.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory] = r
	}

	in.Spec.Volumes = g.ensureVolumes(in.Spec.Volumes)

	return in
}

func (g *Generator) inlineJobPodTemplate(in *corev1.PodTemplateSpec, cmd ...string) (out *corev1.PodTemplateSpec) {
	if len(in.Spec.Containers) == 0 {
		in.Spec.Containers = append(in.Spec.Containers, corev1.Container{Name: phpContainerName})
	}
	ctr := in.Spec.Containers[0]

	g.ensureContainerDefaults(&ctr)
	ctr.Env = []corev1.EnvVar{
		corev1.EnvVar{
			Name: "MY_POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	}

	ctr.VolumeMounts = []corev1.VolumeMount{}
	ctr.Image = "lachlanevenson/k8s-kubectl:v1.10.5"
	ctr.ImagePullPolicy = corev1.PullIfNotPresent

	// find pod candidates using these labels
	l := g.Labels()
	l["app.kubernetes.io/component"] = "web"
	l["app.kubernetes.io/tier"] = "front"

	getPodCmd := fmt.Sprintf("kubectl get pod --namespace $MY_POD_NAMESPACE --sort-by=.status.startTime --field-selector=status.phase==Running -o name -l %s | cut -d'/' -f2 | tail -n1", l)
	shCmd := fmt.Sprintf("kubectl exec $(%s) -c wordpress -- %s", getPodCmd, shellquote.Join(cmd...))
	ctr.Command = []string{"/bin/sh"}
	ctr.Args = []string{"-c", shCmd}

	in.Spec.Containers[0] = ctr

	in.Spec.Volumes = []corev1.Volume{}

	return in
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

	if r, ok := g.WP.Spec.Resources.Requests[wpapi.ResourcePHPWorkers]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "PHP_WORKERS",
			Value: fmt.Sprintf("%d", w),
		})
	}

	if r, ok := g.WP.Spec.Resources.Limits[wpapi.ResourcePHPWorkers]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "PHP_MAX_WORKERS",
			Value: fmt.Sprintf("%d", w),
		})
	}

	if r, ok := g.WP.Spec.Resources.Requests[wpapi.ResourcePHPWorkerMemory]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "PHP_MEMORY_SOFT_LIMIT",
			Value: fmt.Sprintf("%d", w/1024/1024),
		})
	}
	if r, ok := g.WP.Spec.Resources.Limits[wpapi.ResourcePHPWorkerMemory]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "PHP_MEMORY_LIMIT",
			Value: fmt.Sprintf("%d", w/1024/1024),
		})
	}

	if r, ok := g.WP.Spec.Resources.Requests[wpapi.ResourcePHPMaxExecutionSeconds]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "PHP_MAX_EXECUTION_TIME",
			Value: fmt.Sprintf("%d", w),
		})
	}
	if r, ok := g.WP.Spec.Resources.Limits[wpapi.ResourcePHPMaxExecutionSeconds]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "PHP_REQUEST_TIMEOUT",
			Value: fmt.Sprintf("%d", w),
		})
	}

	if r, ok := g.WP.Spec.Resources.Limits[wpapi.ResourceIngressBodySize]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "PHP_MAX_BODY_SIZE",
			Value: fmt.Sprintf("%d", w/1024/1024),
		})
	}

	ctr.Env = core_util.UpsertEnvVars(ctr.Env, env...)
}

func (g *Generator) ensureTempVolumeMounts(ctr *corev1.Container) {
	for _, p := range tempMountPaths {
		ctr.VolumeMounts = core_util.UpsertVolumeMountByPath(ctr.VolumeMounts, corev1.VolumeMount{
			Name:      tempVolumeName,
			ReadOnly:  false,
			MountPath: fmt.Sprintf("/%s", p),
			SubPath:   fmt.Sprintf("%s", p),
		})
	}
}

func (g *Generator) ensureWordpressVolumeMounts(ctr *corev1.Container) {
	ro := g.WP.Spec.ContentVolumeSpec.ReadOnly != nil && *g.WP.Spec.ContentVolumeSpec.ReadOnly

	ctr.VolumeMounts = core_util.UpsertVolumeMountByPath(ctr.VolumeMounts, corev1.VolumeMount{
		Name:      contentVolumeName,
		ReadOnly:  ro,
		MountPath: "/content",
	})

	for _, p := range contentMountPaths {
		ctr.VolumeMounts = core_util.UpsertVolumeMountByPath(ctr.VolumeMounts, corev1.VolumeMount{
			Name:      contentVolumeName,
			ReadOnly:  ro,
			MountPath: fmt.Sprintf("/www/wp-content/%s", p),
			SubPath:   fmt.Sprintf("wp-content/%s", p),
		})
	}

	if g.WP.Spec.MediaVolumeSpec == nil {
		// Mount /wp-content/uploads from content volume if MediaVolumeSpec is
		// not specified
		ctr.VolumeMounts = core_util.UpsertVolumeMountByPath(ctr.VolumeMounts, corev1.VolumeMount{
			Name:      contentVolumeName,
			ReadOnly:  ro,
			MountPath: "/www/wp-content/uploads",
			SubPath:   "wp-content/uploads",
		})
	} else {
		ctr.VolumeMounts = core_util.UpsertVolumeMountByPath(ctr.VolumeMounts, corev1.VolumeMount{
			Name:      mediaVolumeName,
			MountPath: "/www/wp-content/uploads",
			ReadOnly:  g.WP.Spec.MediaVolumeSpec.ReadOnly != nil && *g.WP.Spec.MediaVolumeSpec.ReadOnly,
		})
	}
}

func (g *Generator) ensureVolumes(in []corev1.Volume) []corev1.Volume {
	var v corev1.Volume

	// temporary files
	v = corev1.Volume{Name: tempVolumeName}
	v.VolumeSource = corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{
			Medium:    corev1.StorageMediumMemory,
			SizeLimit: resource.NewQuantity(tempVolumeSize, resource.BinarySI),
		},
	}
	in = core_util.UpsertVolume(in, v)

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
	ctr.Resources.Limits = make(corev1.ResourceList)
	ctr.Resources.Requests = make(corev1.ResourceList)
}

func (g *Generator) ensurePHPContainer(in []corev1.Container) []corev1.Container {
	fi := -1
	for i, container := range in {
		if container.Name == phpContainerName {
			fi = i
			break
		}
	}
	if fi == -1 {
		in = append(in, corev1.Container{Name: phpContainerName})
		fi = len(in) - 1
	}

	// make sure that PHP container is the first in list
	if fi != 0 {
		in[fi], in[0] = in[0], in[fi]
		fi = 0
	}

	ctr := in[fi]
	g.ensureContainerDefaults(&ctr)
	g.ensureWordpressEnv(&ctr)
	g.ensureWordpressVolumeMounts(&ctr)
	g.ensureTempVolumeMounts(&ctr)

	// fill the PHP container spec
	if len(g.WP.Spec.Image) > 0 {
		ctr.Image = g.WP.Spec.Image
	} else {
		ctr.Image = g.DefaultRuntimeImage
	}
	if len(g.WP.Spec.ImagePullPolicy) > 0 {
		ctr.ImagePullPolicy = g.WP.Spec.ImagePullPolicy
	}
	ctr.Args = []string{"/usr/local/bin/php-fpm"}

	if r, ok := g.WP.Spec.Resources.Limits[wpapi.ResourcePHPCPU]; ok {
		ctr.Resources.Limits[corev1.ResourceCPU] = r
	}
	if r, ok := g.WP.Spec.Resources.Limits[wpapi.ResourcePHPMemory]; ok {
		ctr.Resources.Limits[corev1.ResourceMemory] = r
	}
	if r, ok := g.WP.Spec.Resources.Requests[wpapi.ResourcePHPCPU]; ok {
		ctr.Resources.Requests[corev1.ResourceCPU] = r
	}
	if r, ok := g.WP.Spec.Resources.Requests[wpapi.ResourcePHPMemory]; ok {
		ctr.Resources.Requests[corev1.ResourceMemory] = r
	}

	in[fi] = ctr
	return in
}

func (g *Generator) ensureNginxContainer(in []corev1.Container) []corev1.Container {
	fi := -1
	for i, container := range in {
		if container.Name == nginxContainerName {
			fi = i
			break
		}
	}
	if fi == -1 {
		in = append(in, corev1.Container{Name: nginxContainerName})
		fi = len(in) - 1
	}
	ctr := in[fi]

	g.ensureContainerDefaults(&ctr)
	g.ensureWordpressVolumeMounts(&ctr)
	g.ensureTempVolumeMounts(&ctr)

	// fill in the nginx container spec
	if len(g.WP.Spec.Image) > 0 {
		ctr.Image = g.WP.Spec.Image
	} else {
		ctr.Image = g.DefaultRuntimeImage
	}
	if len(g.WP.Spec.ImagePullPolicy) > 0 {
		ctr.ImagePullPolicy = g.WP.Spec.ImagePullPolicy
	}
	ctr.Args = []string{"/usr/local/bin/nginx"}

	env := []corev1.EnvVar{}
	if r, ok := g.WP.Spec.Resources.Limits[wpapi.ResourceIngressBodySize]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "CLIENT_MAX_BODY_SIZE",
			Value: fmt.Sprintf("%d", w/1024/1024),
		})
	}
	ctr.Env = core_util.UpsertEnvVars(ctr.Env, env...)

	if r, ok := g.WP.Spec.Resources.Limits[wpapi.ResourceNginxCPU]; ok {
		ctr.Resources.Limits[corev1.ResourceCPU] = r
	}
	if r, ok := g.WP.Spec.Resources.Limits[wpapi.ResourceNginxMemory]; ok {
		ctr.Resources.Limits[corev1.ResourceMemory] = r
	}
	if r, ok := g.WP.Spec.Resources.Requests[wpapi.ResourceNginxCPU]; ok {
		ctr.Resources.Requests[corev1.ResourceCPU] = r
	}
	if r, ok := g.WP.Spec.Resources.Requests[wpapi.ResourceNginxMemory]; ok {
		ctr.Resources.Requests[corev1.ResourceMemory] = r
	}

	ctr.Ports = []corev1.ContainerPort{
		corev1.ContainerPort{
			Name:          "http",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: 80,
		},
		corev1.ContainerPort{
			Name:          "prometheus",
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: 10080,
		},
	}

	in[fi] = ctr

	return in
}
