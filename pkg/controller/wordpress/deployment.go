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

	apps_util "github.com/appscode/kutil/apps/v1"
	core_util "github.com/appscode/kutil/core/v1"
	"github.com/golang/glog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	wpapi "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
)

const (
	phpContainerName   = "wordpress"
	nginxContainerName = "nginx"
	deploymentName     = "%s"
	contentVolumeName  = "content"
	mediaVolumeName    = "media"
	tempVolumeName     = "temp"
	tempVolumeSize     = 128 * 1024 * 1024
)

var contentMountPaths = []string{"root", "themes", "plugins", "mu-plugins", "languages", "upgrade"}
var tempMountPaths = []string{"run", "tmp"}

func (c *Controller) podLabels(wp *wpapi.Wordpress) labels.Set {
	l := c.instanceLabels(wp)
	l["apps.kubernetes.io/tier"] = "front"
	return l
}

func (c *Controller) syncDeployment(wp *wpapi.Wordpress) error {
	glog.Infof("Syncing deployment for %s/%s", wp.ObjectMeta.Namespace, wp.ObjectMeta.Name)

	wpdef := wp.WithDefaults()
	meta := c.objectMeta(wp, deploymentName)
	meta.Labels["app.kubernetes.io/tier"] = "front"

	_, _, err := apps_util.CreateOrPatchDeployment(c.KubeClient, meta, func(in *appsv1.Deployment) *appsv1.Deployment {
		in.ObjectMeta = c.ensureControllerReference(in.ObjectMeta, wp)

		in.Spec.Selector = metav1.SetAsLabelSelector(c.podLabels(wp))
		in.Spec.Template.ObjectMeta.Labels = c.podLabels(wp)

		if wp.Spec.Replicas != nil {
			in.Spec.Replicas = wp.Spec.Replicas
		}

		c.ensureNginxContainer(wpdef, in)
		c.ensurePHPContainer(wpdef, in)

		in.Spec.Template.Spec.NodeSelector = wpdef.Spec.NodeSelector
		in.Spec.Template.Spec.Tolerations = wpdef.Spec.Tolerations
		in.Spec.Template.Spec.Affinity = &wpdef.Spec.Affinity

		in.Spec.Template.Spec.Volumes = c.ensureVolumes(wpdef, in.Spec.Template.Spec.Volumes)

		return in
	})
	return err
}

func (c *Controller) ensureWordpressEnv(wp *wpapi.Wordpress, ctr *corev1.Container) {
	ctr.Env = core_util.UpsertEnvVars(ctr.Env, wp.Spec.Env...)

	var domains []string
	for _, d := range wp.Spec.Domains {
		domains = append(domains, string(d))
	}
	env := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "WORDPRESS_REPO_URL",
			Value: string(wp.Spec.RepoURL),
		},
		corev1.EnvVar{
			Name:  "WORDPRESS_REPO_REF",
			Value: wp.Spec.RepoRef,
		},
		corev1.EnvVar{
			Name:  "WORDPRESS_DOMAINS",
			Value: strings.Join(domains, ","),
		},
	}

	if r, ok := wp.Spec.Resources.Requests[wpapi.ResourcePHPWorkers]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "PHP_WORKERS",
			Value: fmt.Sprintf("%d", w),
		})
	}

	if r, ok := wp.Spec.Resources.Limits[wpapi.ResourcePHPWorkers]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "PHP_MAX_WORKERS",
			Value: fmt.Sprintf("%d", w),
		})
	}

	if r, ok := wp.Spec.Resources.Requests[wpapi.ResourcePHPWorkerMemory]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "PHP_MEMORY_SOFT_LIMIT",
			Value: fmt.Sprintf("%d", w/1024/1024),
		})
	}
	if r, ok := wp.Spec.Resources.Limits[wpapi.ResourcePHPWorkerMemory]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "PHP_MEMORY_LIMIT",
			Value: fmt.Sprintf("%d", w/1024/1024),
		})
	}

	if r, ok := wp.Spec.Resources.Requests[wpapi.ResourcePHPMaxExecutionSeconds]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "PHP_MAX_EXECUTION_TIME",
			Value: fmt.Sprintf("%d", w),
		})
	}
	if r, ok := wp.Spec.Resources.Limits[wpapi.ResourcePHPMaxExecutionSeconds]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "PHP_REQUEST_TIMEOUT",
			Value: fmt.Sprintf("%d", w),
		})
	}

	if r, ok := wp.Spec.Resources.Limits[wpapi.ResourceIngressBodySize]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "PHP_MAX_BODY_SIZE",
			Value: fmt.Sprintf("%d", w/1024/1024),
		})
	}

	ctr.Env = core_util.UpsertEnvVars(ctr.Env, env...)
}

func (c *Controller) ensureContainerDefaults(ctr *corev1.Container) {
	ctr.Resources.Limits = make(corev1.ResourceList)
	ctr.Resources.Requests = make(corev1.ResourceList)
}

func (c *Controller) ensurePHPContainer(wp *wpapi.Wordpress, in *appsv1.Deployment) {
	fi := -1
	for i, container := range in.Spec.Template.Spec.Containers {
		if container.Name == phpContainerName {
			fi = i
			break
		}
	}
	if fi == -1 {
		in.Spec.Template.Spec.Containers = append(in.Spec.Template.Spec.Containers, corev1.Container{Name: phpContainerName})
		fi = len(in.Spec.Template.Spec.Containers) - 1
	}

	// // make sure that PHP container is the first in list
	if fi != 0 {
		in.Spec.Template.Spec.Containers[fi], in.Spec.Template.Spec.Containers[0] = in.Spec.Template.Spec.Containers[0], in.Spec.Template.Spec.Containers[fi]
		fi = 0
	}
	ctr := in.Spec.Template.Spec.Containers[fi]
	c.ensureContainerDefaults(&ctr)
	c.ensureWordpressEnv(wp, &ctr)
	c.ensureWordpressVolumeMounts(wp, &ctr)
	c.ensureTempVolumeMounts(wp, &ctr)

	// fill the PHP container spec
	if len(wp.Spec.Image) > 0 {
		ctr.Image = wp.Spec.Image
	} else {
		ctr.Image = c.RuntimeImage
	}
	ctr.Args = []string{"/usr/local/bin/php-fpm"}

	if r, ok := wp.Spec.Resources.Limits[wpapi.ResourcePHPCPU]; ok {
		ctr.Resources.Limits[corev1.ResourceCPU] = r
	}
	if r, ok := wp.Spec.Resources.Limits[wpapi.ResourcePHPMemory]; ok {
		ctr.Resources.Limits[corev1.ResourceMemory] = r
	}
	if r, ok := wp.Spec.Resources.Requests[wpapi.ResourcePHPCPU]; ok {
		ctr.Resources.Requests[corev1.ResourceCPU] = r
	}
	if r, ok := wp.Spec.Resources.Requests[wpapi.ResourcePHPMemory]; ok {
		ctr.Resources.Requests[corev1.ResourceMemory] = r
	}

	in.Spec.Template.Spec.Containers[fi] = ctr
}

func (c *Controller) ensureVolumes(wp *wpapi.Wordpress, in []corev1.Volume) []corev1.Volume {
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

	v = corev1.Volume{Name: contentVolumeName}
	if wp.Spec.ContentVolumeSpec.PersistentVolumeClaim != nil {
		v.VolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: fmt.Sprintf(contentPVCName, wp.Name),
			},
		}
	} else if wp.Spec.ContentVolumeSpec.HostPath != nil {
		v.VolumeSource = corev1.VolumeSource{HostPath: wp.Spec.ContentVolumeSpec.HostPath}
	} else {
		d := wp.Spec.ContentVolumeSpec.EmptyDir
		if d == nil {
			d = &corev1.EmptyDirVolumeSource{}
		}
		v.VolumeSource = corev1.VolumeSource{EmptyDir: d}
	}
	in = core_util.UpsertVolume(in, v)

	if wp.Spec.MediaVolumeSpec == nil {
		in = core_util.EnsureVolumeDeleted(in, mediaVolumeName)
		// Return is MediaVolumeSpec is not defined
		return in
	}

	v = corev1.Volume{Name: mediaVolumeName}
	if wp.Spec.MediaVolumeSpec.PersistentVolumeClaim != nil {
		v.VolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: fmt.Sprintf(contentPVCName, wp.Name),
			},
		}
	} else if wp.Spec.MediaVolumeSpec.HostPath != nil {
		v.VolumeSource = corev1.VolumeSource{HostPath: wp.Spec.MediaVolumeSpec.HostPath}
	} else {
		d := wp.Spec.MediaVolumeSpec.EmptyDir
		if d == nil {
			d = &corev1.EmptyDirVolumeSource{}
		}
		v.VolumeSource = corev1.VolumeSource{EmptyDir: d}
	}
	in = core_util.UpsertVolume(in, v)
	return in
}

func (c *Controller) ensureTempVolumeMounts(wp *wpapi.Wordpress, ctr *corev1.Container) {
	for _, p := range tempMountPaths {
		ctr.VolumeMounts = core_util.UpsertVolumeMountByPath(ctr.VolumeMounts, corev1.VolumeMount{
			Name:      tempVolumeName,
			ReadOnly:  false,
			MountPath: fmt.Sprintf("/%s", p),
			SubPath:   fmt.Sprintf("%s", p),
		})
	}
}

func (c *Controller) ensureWordpressVolumeMounts(wp *wpapi.Wordpress, ctr *corev1.Container) {
	ro := wp.Spec.ContentVolumeSpec.ReadOnly != nil && *wp.Spec.ContentVolumeSpec.ReadOnly

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

	if wp.Spec.MediaVolumeSpec == nil {
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
			ReadOnly:  wp.Spec.MediaVolumeSpec.ReadOnly != nil && *wp.Spec.MediaVolumeSpec.ReadOnly,
		})
	}
}

func (c *Controller) ensureNginxContainer(wp *wpapi.Wordpress, in *appsv1.Deployment) {
	fi := -1
	for i, container := range in.Spec.Template.Spec.Containers {
		if container.Name == nginxContainerName {
			fi = i
			break
		}
	}
	if fi == -1 {
		in.Spec.Template.Spec.Containers = append(in.Spec.Template.Spec.Containers, corev1.Container{Name: nginxContainerName})
		fi = len(in.Spec.Template.Spec.Containers) - 1
	}
	ctr := in.Spec.Template.Spec.Containers[fi]

	c.ensureContainerDefaults(&ctr)
	c.ensureWordpressVolumeMounts(wp, &ctr)
	c.ensureTempVolumeMounts(wp, &ctr)

	// fill in the nginx container spec
	if len(wp.Spec.Image) > 0 {
		ctr.Image = wp.Spec.Image
	} else {
		ctr.Image = c.RuntimeImage
	}
	ctr.Args = []string{"/usr/local/bin/nginx"}

	env := []corev1.EnvVar{}
	if r, ok := wp.Spec.Resources.Limits[wpapi.ResourceIngressBodySize]; ok {
		w, _ := r.AsInt64()
		env = append(env, corev1.EnvVar{
			Name:  "CLIENT_MAX_BODY_SIZE",
			Value: fmt.Sprintf("%d", w/1024/1024),
		})
	}
	ctr.Env = core_util.UpsertEnvVars(ctr.Env, env...)

	if r, ok := wp.Spec.Resources.Limits[wpapi.ResourceNginxCPU]; ok {
		ctr.Resources.Limits[corev1.ResourceCPU] = r
	}
	if r, ok := wp.Spec.Resources.Limits[wpapi.ResourceNginxMemory]; ok {
		ctr.Resources.Limits[corev1.ResourceMemory] = r
	}
	if r, ok := wp.Spec.Resources.Requests[wpapi.ResourceNginxCPU]; ok {
		ctr.Resources.Requests[corev1.ResourceCPU] = r
	}
	if r, ok := wp.Spec.Resources.Requests[wpapi.ResourceNginxMemory]; ok {
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

	in.Spec.Template.Spec.Containers[fi] = ctr
}
