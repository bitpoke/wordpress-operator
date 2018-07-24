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
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	wpCronName       = "%s-wp-cron"
	deploymentName   = "%s"
	serviceName      = "%s"
	ingressName      = "%s"
	webrootPVCName   = "%s"
	mediaPVCName     = "%s-media"
	dbUpgradeJobName = "%s-db-upgrade"

	webrootVolumeName = "webroot"
	mediaVolumeName   = "media"
)

var (
	DefaultWebImage string = "docker.io/library/wordpress:latest"
	DefaultCLIImage string = "docker.io/library/wordpress:cli"

	oneReplica int32 = 1
)

func (wp *Wordpress) GetWebrootPVCName() string { return fmt.Sprintf(webrootPVCName, wp.Name) }
func (wp *Wordpress) GetWPCronName() string     { return fmt.Sprintf(wpCronName, wp.Name) }
func (wp *Wordpress) GetDeploymentName() string { return fmt.Sprintf(deploymentName, wp.Name) }
func (wp *Wordpress) GetIngressName() string    { return fmt.Sprintf(ingressName, wp.Name) }
func (wp *Wordpress) GetMediaPVCName() string   { return fmt.Sprintf(mediaPVCName, wp.Name) }
func (wp *Wordpress) GetServiceName() string    { return fmt.Sprintf(serviceName, wp.Name) }
func (wp *Wordpress) GetDBUpgradeJobName() string {
	ver := wp.GetWPVersion()
	prefix := fmt.Sprintf("%s-%x", wp.Name, md5.Sum([]byte(ver)))
	return fmt.Sprintf(dbUpgradeJobName, prefix)
}
func (wp *Wordpress) GetWPVersion() string {
	var images []string

	for _, c := range wp.Spec.WebPodTemplate.Spec.Containers {
		images = append(images, c.Image)
	}

	return strings.Join(images, ",")
}

// WithDefaults returns a Wordpress object with defaults filled in
// Controller should always apply defaults before dowing work
func (wp *Wordpress) WithDefaults() (d *Wordpress) {
	d = wp

	if d.Spec.Replicas == nil || *d.Spec.Replicas < 1 {
		d.Spec.Replicas = &oneReplica
	}

	if len(d.Spec.VolumeMountsSpec) == 0 {
		d.Spec.VolumeMountsSpec = []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      webrootVolumeName,
				MountPath: "/var/www/html",
			},
		}
		if d.Spec.MediaVolumeSpec != nil {
			d.Spec.VolumeMountsSpec = append(d.Spec.VolumeMountsSpec, corev1.VolumeMount{
				Name:      mediaVolumeName,
				MountPath: "/var/www/html/wp-content/uploads",
			})
		}
	}

	if d.Spec.WebPodTemplate == nil {
		d.Spec.WebPodTemplate = &corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:  "wordpress",
						Image: DefaultWebImage,
						Ports: []corev1.ContainerPort{
							corev1.ContainerPort{
								Name:          "http",
								ContainerPort: 80,
								Protocol:      corev1.ProtocolTCP,
							},
						},
					},
				},
			},
		}
	}

	if d.Spec.CLIPodTemplate == nil {
		d.Spec.CLIPodTemplate = &corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:  "wp-cli",
						Image: DefaultCLIImage,
					},
				},
			},
		}
	}

	if d.Spec.ServiceSpec == nil {
		d.Spec.ServiceSpec = &corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromString("http"),
				},
			},
		}
	}

	return
}

// LabelsSet returns a general label set to apply to objects, relative to the
// Wordpress API object
func (wp *Wordpress) LabelsSet() labels.Set {
	return labels.Set{
		"app.kubernetes.io/name":           "wordpress",
		"app.kubernetes.io/app-instance":   wp.Name,
		"app.kubernetes.io/deploy-manager": "wordpress-operator",
	}
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
func (wp *Wordpress) WebPodTemplateSpec() (out *corev1.PodTemplateSpec) {
	if wp.Spec.WebPodTemplate != nil {
		out = wp.Spec.WebPodTemplate.DeepCopy()
	} else {
		out = &corev1.PodTemplateSpec{}
	}

	out.ObjectMeta.Labels = wp.WebPodLabels()

	out.Spec.Volumes = wp.ensureWordpressVolumes(out.Spec.Volumes)

	for i := range out.Spec.InitContainers {
		wp.ensureWordpressEnv(&out.Spec.InitContainers[i])
		wp.ensureWordpressVolumeMounts(&out.Spec.InitContainers[i])
	}
	for i := range out.Spec.Containers {
		wp.ensureWordpressEnv(&out.Spec.Containers[i])
		wp.ensureWordpressVolumeMounts(&out.Spec.Containers[i])
	}

	return
}

// JobPodTemplate generates a pod template spec suitable for WP CLI background
// jobs
func (wp *Wordpress) JobPodTemplateSpec(cmd ...string) (out *corev1.PodTemplateSpec) {
	if wp.Spec.WebPodTemplate != nil {
		out = wp.Spec.CLIPodTemplate.DeepCopy()
	} else {
		out = &corev1.PodTemplateSpec{}
	}

	out.ObjectMeta.Labels = wp.WebPodLabels()

	out.Spec.Volumes = wp.ensureWordpressVolumes(out.Spec.Volumes)
	out.Spec.RestartPolicy = corev1.RestartPolicyNever

	for i := range out.Spec.InitContainers {
		wp.ensureWordpressEnv(&out.Spec.InitContainers[i])
		wp.ensureWordpressVolumeMounts(&out.Spec.InitContainers[i])
	}
	for i := range out.Spec.Containers {
		wp.ensureWordpressEnv(&out.Spec.Containers[i])
		wp.ensureWordpressVolumeMounts(&out.Spec.Containers[i])
	}

	for i, c := range out.Spec.Containers {
		if c.Name == "wp-cli" {
			out.Spec.Containers[i].Args = append(out.Spec.Containers[i].Args, cmd...)
		}
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

func (wp *Wordpress) ensureWordpressVolumes(in []corev1.Volume) []corev1.Volume {
	var v corev1.Volume

	// content (plugins, themes, etc.)
	v = corev1.Volume{Name: webrootVolumeName}
	if wp.Spec.WebrootVolumeSpec.PersistentVolumeClaim != nil {
		v.VolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: wp.GetWebrootPVCName(),
			},
		}
	} else if wp.Spec.WebrootVolumeSpec.HostPath != nil {
		v.VolumeSource = corev1.VolumeSource{HostPath: wp.Spec.WebrootVolumeSpec.HostPath}
	} else {
		d := wp.Spec.WebrootVolumeSpec.EmptyDir
		if d == nil {
			d = &corev1.EmptyDirVolumeSource{}
		}
		v.VolumeSource = corev1.VolumeSource{EmptyDir: d}
	}
	in = core_util.UpsertVolume(in, v)

	// media files
	if wp.Spec.MediaVolumeSpec == nil {
		in = core_util.EnsureVolumeDeleted(in, mediaVolumeName)
		// Return is MediaVolumeSpec is not defined
		return in
	}

	v = corev1.Volume{Name: mediaVolumeName}
	if wp.Spec.MediaVolumeSpec.PersistentVolumeClaim != nil {
		v.VolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: wp.GetMediaPVCName(),
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

func (wp *Wordpress) ensureContainerDefaults(ctr *corev1.Container) {
	if len(ctr.Resources.Limits) == 0 {
		ctr.Resources.Limits = make(corev1.ResourceList)
	}
	if len(ctr.Resources.Requests) == 0 {
		ctr.Resources.Requests = make(corev1.ResourceList)
	}
}
