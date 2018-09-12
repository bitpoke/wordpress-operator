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
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/presslabs/controller-util/syncer"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
)

// NewWPCronSyncer returns a new sync.Interface for reconciling wp-cron CronJob
func NewWPCronSyncer(wp *wordpressv1alpha1.Wordpress, rt *wordpressv1alpha1.WordpressRuntime) syncer.Interface {
	obj := &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wp.GetWPCronName(),
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

	return syncer.New("WPCron", wp, obj, func(existing runtime.Object) error {
		out := existing.(*batchv1beta1.CronJob)

		out.Labels = wp.LabelsForComponent("wp-cron")

		out.Spec.Schedule = "* * * * *"
		out.Spec.ConcurrencyPolicy = "Forbid"
		out.Spec.StartingDeadlineSeconds = &cronStartingDeadlineSeconds
		out.Spec.SuccessfulJobsHistoryLimit = &successfulJobsHistoryLimit
		out.Spec.FailedJobsHistoryLimit = &failedJobsHistoryLimit

		out.Spec.JobTemplate.ObjectMeta.Labels = wp.LabelsForComponent("wp-cron")
		out.Spec.JobTemplate.Spec.BackoffLimit = &backoffLimit
		out.Spec.JobTemplate.Spec.ActiveDeadlineSeconds = &activeDeadlineSeconds

		cmd := []string{"wp", "cron", "event", "run", "--due-now"}
		out.Spec.JobTemplate.Spec.Template = *wp.JobPodTemplateSpec(rt, cmd...)
		out.Spec.JobTemplate.Spec.Template.ObjectMeta.Labels = wp.LabelsForComponent("wp-cron")

		return nil
	})
}
