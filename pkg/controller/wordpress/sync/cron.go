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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
)

const (
	// EventReasonWPCronFailed is the event reason for a failed wp-cron CronJob reconcile
	EventReasonWPCronFailed EventReason = "WPCronFailed"
	// EventReasonWPCronUpdated is the event reason for a successful wp-cron CronJob reconcile
	EventReasonWPCronUpdated EventReason = "WPCronUpdated"
)

type wpCronSyncer struct {
	scheme   *runtime.Scheme
	wp       *wordpressv1alpha1.Wordpress
	rt       *wordpressv1alpha1.WordpressRuntime
	key      types.NamespacedName
	existing *batchv1beta1.CronJob
}

// NewWPCronSyncer returns a new sync.Interface for reconciling wp-cron CronJob
func NewWPCronSyncer(wp *wordpressv1alpha1.Wordpress, rt *wordpressv1alpha1.WordpressRuntime, r *runtime.Scheme) Interface {
	return &wpCronSyncer{
		scheme:   r,
		wp:       wp,
		rt:       rt,
		existing: &batchv1beta1.CronJob{},
		key: types.NamespacedName{
			Name:      wp.GetWPCronName(),
			Namespace: wp.Namespace,
		},
	}
}

func (s *wpCronSyncer) GetKey() types.NamespacedName                 { return s.key }
func (s *wpCronSyncer) GetExistingObjectPlaceholder() runtime.Object { return s.existing }

func (s *wpCronSyncer) T(in runtime.Object) (runtime.Object, error) {
	var (
		cronStartingDeadlineSeconds int64 = 10
		backoffLimit                int32
		activeDeadlineSeconds       int64 = 10
		successfulJobsHistoryLimit  int32 = 3
		failedJobsHistoryLimit      int32 = 1
	)
	out := in.(*batchv1beta1.CronJob)

	out.Name = s.key.Name
	out.Namespace = s.key.Namespace
	out.Labels = s.wp.LabelsForComponent("wp-cron")
	if err := controllerutil.SetControllerReference(s.wp, out, s.scheme); err != nil {
		return nil, err
	}

	out.Spec.Schedule = "* * * * *"
	out.Spec.ConcurrencyPolicy = "Forbid"
	out.Spec.StartingDeadlineSeconds = &cronStartingDeadlineSeconds
	out.Spec.SuccessfulJobsHistoryLimit = &successfulJobsHistoryLimit
	out.Spec.FailedJobsHistoryLimit = &failedJobsHistoryLimit

	out.Spec.JobTemplate.ObjectMeta.Labels = s.wp.LabelsForComponent("wp-cron")
	out.Spec.JobTemplate.Spec.BackoffLimit = &backoffLimit
	out.Spec.JobTemplate.Spec.ActiveDeadlineSeconds = &activeDeadlineSeconds

	cmd := []string{"wp", "cron", "event", "run", "--due-now"}
	out.Spec.JobTemplate.Spec.Template = *s.wp.JobPodTemplateSpec(s.rt, cmd...)
	out.Spec.JobTemplate.Spec.Template.ObjectMeta.Labels = s.wp.LabelsForComponent("wp-cron")

	return out, nil
}

func (s *wpCronSyncer) GetErrorEventReason(err error) EventReason {
	if err != nil {
		return EventReasonWPCronFailed
	}
	return EventReasonWPCronUpdated
}
