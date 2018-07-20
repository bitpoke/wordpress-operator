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
	batch_util "github.com/appscode/kutil/batch/v1beta1"
	"github.com/golang/glog"
	batchv1beta1 "k8s.io/api/batch/v1beta1"

	wpapi "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	"github.com/presslabs/wordpress-operator/pkg/factory/wordpress"
)

var (
	cronStartingDeadlineSeconds int64 = 10
)

func (c *Controller) syncCron(wp *wpapi.Wordpress) error {
	glog.Infof("Syncing wp-cron for %s/%s", wp.ObjectMeta.Namespace, wp.ObjectMeta.Name)

	wpf := wordpress.Generator{WP: wp}

	labels := wpf.Labels()
	labels["app.kubernetes.io/component"] = "wp-cron"

	meta := c.objectMeta(wp, wpf.WPCronName())
	meta.Labels = labels

	var backoffLimit int32 = 0
	var activeDeadlineSeconds int64 = 10

	_, _, err := batch_util.CreateOrPatchCronJob(c.KubeClient, meta, func(in *batchv1beta1.CronJob) *batchv1beta1.CronJob {
		in.ObjectMeta = c.ensureControllerReference(in.ObjectMeta, wp)

		in.Spec.Schedule = "* * * * *"
		in.Spec.ConcurrencyPolicy = "Forbid"
		in.Spec.StartingDeadlineSeconds = &cronStartingDeadlineSeconds

		in.Spec.JobTemplate.ObjectMeta.Labels = labels
		in.Spec.JobTemplate.Spec.BackoffLimit = &backoffLimit
		in.Spec.JobTemplate.Spec.ActiveDeadlineSeconds = &activeDeadlineSeconds

		cmd := []string{"wp", "cron", "event", "run", "--due-now"}
		// cmd := []string{"sleep", "15"}
		in.Spec.JobTemplate.Spec.Template = *wpf.JobPodTemplateSpec(&in.Spec.JobTemplate.Spec.Template, cmd...)
		in.Spec.JobTemplate.Spec.Template.ObjectMeta.Labels = labels

		return in
	})
	return err
}
