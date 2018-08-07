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
package v1alpha1

import (
	"crypto/md5"
	"fmt"
	"strings"

	core_util "github.com/appscode/kutil/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	wpCronName       = "%s-wp-cron"
	deploymentName   = "%s"
	serviceName      = "%s"
	ingressName      = "%s"
	webrootPVCName   = "%s-webroot"
	mediaPVCName     = "%s-media"
	dbUpgradeJobName = "%s-db-upgrade"

	webrootVolumeName = "webroot"
	mediaVolumeName   = "media"
)

var (
	defaultImage       = "defaultImage"
	oneReplica   int32 = 1
)

func (wp *Wordpress) GetWebrootPVCName(rt *WordpressRuntime) string {
	return fmt.Sprintf(webrootPVCName, wp.Name)
}
func (wp *Wordpress) GetWPCronName(rt *WordpressRuntime) string {
	return fmt.Sprintf(wpCronName, wp.Name)
}
func (wp *Wordpress) GetDeploymentName(rt *WordpressRuntime) string {
	return fmt.Sprintf(deploymentName, wp.Name)
}
func (wp *Wordpress) GetIngressName(rt *WordpressRuntime) string {
	return fmt.Sprintf(ingressName, wp.Name)
}
func (wp *Wordpress) GetMediaPVCName(rt *WordpressRuntime) string {
	return fmt.Sprintf(mediaPVCName, wp.Name)
}
func (wp *Wordpress) GetServiceName(rt *WordpressRuntime) string {
	return fmt.Sprintf(serviceName, wp.Name)
}
func (wp *Wordpress) GetDBUpgradeJobName(rt *WordpressRuntime) string {
	ver := wp.GetWPVersion(rt)
	prefix := fmt.Sprintf("%s-%x", wp.Name, md5.Sum([]byte(ver)))
	return fmt.Sprintf(dbUpgradeJobName, prefix)
}
func (wp *Wordpress) GetWPVersion(rt *WordpressRuntime) string {
	image := rt.Spec.DefaultImage
	if len(wp.Spec.Image) > 0 {
		image = wp.Spec.Image
	}
	return image
}

// SetDefaults mutates a Wordpress object and sets default values
// Controller should always apply defaults before passing it down to workers
func (wp *Wordpress) SetDefaults() {
	if wp.Spec.Replicas == nil || *wp.Spec.Replicas < 1 {
		wp.Spec.Replicas = &oneReplica
	}
}

// LabelsSet returns a general label set to apply to objects, relative to the
// Wordpress API object
func (wp *Wordpress) LabelsSet() labels.Set {
	l := labels.Set{}
	for k, v := range wp.Spec.Labels {
		l[k] = v
	}
	l["app.kubernetes.io/name"] = "wordpress"
	l["app.kubernetes.io/app-instance"] = wp.Name
	l["app.kubernetes.io/deploy-manager"] = "wordpress-operator"

	return l
}

// LabelsForTier returns a label set object with tier label filled in
func (wp *Wordpress) LabelsForTier(tier string) labels.Set {
	l := wp.LabelsSet()
	l["app.kubernetes.io/tier"] = tier
	return l
}

// LabelsForComponent returns a label set object with component label filled in
func (wp *Wordpress) LabelsForComponent(component string) labels.Set {
	l := wp.LabelsSet()
	l["app.kubernetes.io/component"] = component
	return l
}

// WebPodLabels returns the labels suitable Wordpress Web Pods
func (wp *Wordpress) WebPodLabels() labels.Set {
	l := wp.LabelsForTier("front")
	l["app.kubernetes.io/component"] = "web"

	return l
}

// WebPodTemplateSpec generates a pod template spec suitable for use in Wordpress
// deployment
func (wp *Wordpress) WebPodTemplateSpec(rt *WordpressRuntime) (out *corev1.PodTemplateSpec) {
	if rt.Spec.WebPodTemplate != nil {
		out = rt.Spec.WebPodTemplate.DeepCopy()
	} else {
		out = &corev1.PodTemplateSpec{}
	}

	out.ObjectMeta.Labels = wp.WebPodLabels()

	out.Spec.Volumes = wp.ensureWordpressVolumes(out.Spec.Volumes, rt)

	for i := range out.Spec.InitContainers {
		wp.ensureWordpressEnv(&out.Spec.InitContainers[i])
		wp.ensureWordpressVolumeMounts(&out.Spec.InitContainers[i])
		wp.setContainerImage(&out.Spec.InitContainers[i], rt)
	}
	for i := range out.Spec.Containers {
		wp.ensureWordpressEnv(&out.Spec.Containers[i])
		wp.ensureWordpressVolumeMounts(&out.Spec.Containers[i])
		wp.setContainerImage(&out.Spec.Containers[i], rt)
	}

	out.Spec.ImagePullSecrets = append(out.Spec.ImagePullSecrets, wp.Spec.ImagePullSecrets...)
	if len(wp.Spec.ServiceAccountName) > 0 {
		out.Spec.ServiceAccountName = wp.Spec.ServiceAccountName
	}
	return
}

// JobPodTemplate generates a pod template spec suitable for WP CLI background
// jobs
func (wp *Wordpress) JobPodTemplateSpec(rt *WordpressRuntime, cmd ...string) (out *corev1.PodTemplateSpec) {
	if rt.Spec.CLIPodTemplate != nil {
		out = rt.Spec.CLIPodTemplate.DeepCopy()
	} else {
		out = &corev1.PodTemplateSpec{}
	}

	out.ObjectMeta.Labels = wp.WebPodLabels()

	out.Spec.Volumes = wp.ensureWordpressVolumes(out.Spec.Volumes, rt)
	out.Spec.RestartPolicy = corev1.RestartPolicyNever

	for i := range out.Spec.InitContainers {
		wp.ensureWordpressEnv(&out.Spec.InitContainers[i])
		wp.ensureWordpressVolumeMounts(&out.Spec.InitContainers[i])
		wp.setContainerImage(&out.Spec.InitContainers[i], rt)
	}
	for i := range out.Spec.Containers {
		wp.ensureWordpressEnv(&out.Spec.Containers[i])
		wp.ensureWordpressVolumeMounts(&out.Spec.Containers[i])
		wp.setContainerImage(&out.Spec.Containers[i], rt)
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
	return
}

func (wp *Wordpress) ensureWordpressEnv(ctr *corev1.Container) {
	ctr.Env = core_util.UpsertEnvVars(ctr.Env, wp.Spec.Env...)

	var domains []string
	for _, d := range wp.Spec.Domains {
		domains = append(domains, string(d))
	}
	env := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "WORDPRESS_DOMAINS",
			Value: strings.Join(domains, ","),
		},
	}

	ctr.Env = core_util.UpsertEnvVars(ctr.Env, env...)
	ctr.EnvFrom = append(wp.Spec.EnvFrom, ctr.EnvFrom...)
}

func (wp *Wordpress) ensureWordpressVolumeMounts(ctr *corev1.Container) {
	for _, v := range wp.Spec.VolumeMountsSpec {
		ctr.VolumeMounts = core_util.UpsertVolumeMountByPath(ctr.VolumeMounts, v)
	}
}

func ensureVolume(name, pvcName string, volSpec *WordpressVolumeSpec, in []corev1.Volume) []corev1.Volume {
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

func (wp *Wordpress) ensureWordpressVolumes(in []corev1.Volume, rt *WordpressRuntime) []corev1.Volume {
	// webroot (plugins, themes, etc.)
	volSpec := rt.Spec.WebrootVolumeSpec
	if wp.Spec.WebrootVolumeSpec != nil {
		volSpec = wp.Spec.WebrootVolumeSpec
	}
	in = ensureVolume(webrootVolumeName, wp.GetWebrootPVCName(rt), volSpec, in)

	volSpec = rt.Spec.MediaVolumeSpec
	if wp.Spec.MediaVolumeSpec != nil {
		volSpec = wp.Spec.MediaVolumeSpec
	}
	in = ensureVolume(mediaVolumeName, wp.GetMediaPVCName(rt), volSpec, in)

	return in
}

func (wp *Wordpress) setContainerImage(ctr *corev1.Container, rt *WordpressRuntime) {
	if ctr.Image != defaultImage {
		return
	}

	image := rt.Spec.DefaultImage
	if len(wp.Spec.Image) > 0 {
		image = wp.Spec.Image
	}

	imagePullPolicy := rt.Spec.DefaultImagePullPolicy
	if len(wp.Spec.ImagePullPolicy) > 0 {
		imagePullPolicy = wp.Spec.ImagePullPolicy
	}

	ctr.Image = image
	if len(imagePullPolicy) > 0 {
		ctr.ImagePullPolicy = imagePullPolicy
	}
}

func (wp *Wordpress) ensureContainerDefaults(ctr *corev1.Container) {
	if len(ctr.Resources.Limits) == 0 {
		ctr.Resources.Limits = make(corev1.ResourceList)
	}
	if len(ctr.Resources.Requests) == 0 {
		ctr.Resources.Requests = make(corev1.ResourceList)
	}
}
