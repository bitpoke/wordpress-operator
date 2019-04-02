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

package sync

import (
	"fmt"
	"github.com/appscode/mergo"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/presslabs/controller-util/mergo/transformers"
	"github.com/presslabs/controller-util/syncer"

	"github.com/presslabs/wordpress-operator/pkg/internal/wordpress"
)

const curlImage = "buildpack-deps:stretch-curl"

// NewWPCronSyncer returns a new sync.Interface for reconciling wp-cron CronJob
func NewWPCronSyncer(wp *wordpress.Wordpress, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	objLabels := wp.ComponentLabels(wordpress.WordpressCron)

	obj := &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wp.ComponentName(wordpress.WordpressCron),
			Namespace: wp.Namespace,
		},
	}

	var (
		cronStartingDeadlineSeconds int64 = 10
		backoffLimit                int32
		activeDeadlineSeconds       int64 = 10
		successfulJobsHistoryLimit  int32 = 3
		failedJobsHistoryLimit      int32 = 1
	)

	return syncer.NewObjectSyncer("WPCron", wp.Unwrap(), obj, c, scheme, func(existing runtime.Object) error {
		out := existing.(*batchv1beta1.CronJob)

		out.Labels = labels.Merge(labels.Merge(out.Labels, objLabels), controllerLabels)

		out.Spec.Schedule = "* * * * *"
		out.Spec.ConcurrencyPolicy = "Forbid"
		out.Spec.StartingDeadlineSeconds = &cronStartingDeadlineSeconds
		out.Spec.SuccessfulJobsHistoryLimit = &successfulJobsHistoryLimit
		out.Spec.FailedJobsHistoryLimit = &failedJobsHistoryLimit

		out.Spec.JobTemplate.ObjectMeta.Labels = labels.Merge(objLabels, controllerLabels)
		out.Spec.JobTemplate.Spec.BackoffLimit = &backoffLimit
		out.Spec.JobTemplate.Spec.ActiveDeadlineSeconds = &activeDeadlineSeconds

		hostHeader := fmt.Sprintf("Host: %s", wp.Spec.Domains[0])
		svcHostname := fmt.Sprintf("%s.%s.svc", wp.Name, wp.Namespace)
		url := fmt.Sprintf("http://%s/wp/wp-cron.php?doing_wp_cron", svcHostname)

		// curl -s -I --max-time 30 -H "Host: <Host>" "http://<site>.<namespace>.svc/wp-cron.php?doing_wp_cron"
		cmd := []string{"curl", "-s", "-I", "--max-time", "30", "-H", hostHeader, url}

		template := corev1.PodTemplateSpec{}
		template.ObjectMeta.Labels = wp.JobPodLabels()
		template.Spec.Containers = []corev1.Container{
			{
				Name:  "curl",
				Image: curlImage,
				Args:  cmd,
			},
		}
		template.Spec.RestartPolicy = corev1.RestartPolicyNever

		out.Spec.JobTemplate.Spec.Template.ObjectMeta = template.ObjectMeta

		err := mergo.Merge(&out.Spec.JobTemplate.Spec.Template.Spec, template.Spec,
			mergo.WithTransformers(transformers.PodSpec))

		if err != nil {
			return err
		}

		return nil
	})
}
