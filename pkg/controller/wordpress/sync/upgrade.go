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
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
)

const (
	// EventReasonDBUpgradeJobFailed is the event reason for a failed database upgrade Job reconcile
	EventReasonDBUpgradeJobFailed EventReason = "DBUpgradeJobFailed"
	// EventReasonDBUpgradeJobUpdated is the event reason for a successful database upgrade Job reconcile
	EventReasonDBUpgradeJobUpdated EventReason = "DBUpgradeJobUpdated"
)

type dbUpgradeJobSyncer struct {
	scheme   *runtime.Scheme
	wp       *wordpressv1alpha1.Wordpress
	rt       *wordpressv1alpha1.WordpressRuntime
	key      types.NamespacedName
	existing *batchv1.Job
}

// NewDBUpgradeJobSyncer returns a new sync.Interface for reconciling database upgrade Job
func NewDBUpgradeJobSyncer(wp *wordpressv1alpha1.Wordpress, rt *wordpressv1alpha1.WordpressRuntime, r *runtime.Scheme) Interface {
	return &dbUpgradeJobSyncer{
		scheme:   r,
		wp:       wp,
		rt:       rt,
		existing: &batchv1.Job{},
		key: types.NamespacedName{
			Name:      wp.GetDBUpgradeJobName(rt),
			Namespace: wp.Namespace,
		},
	}
}

func (s *dbUpgradeJobSyncer) GetKey() types.NamespacedName                 { return s.key }
func (s *dbUpgradeJobSyncer) GetExistingObjectPlaceholder() runtime.Object { return s.existing }

func (s *dbUpgradeJobSyncer) T(in runtime.Object) (runtime.Object, error) {
	var (
		backoffLimit          int32
		activeDeadlineSeconds int64 = 10
	)
	out := in.(*batchv1.Job)

	if !out.CreationTimestamp.IsZero() {
		// TODO(calind): handle the case that the existing job is failed
		return out, nil
	}

	image := s.wp.GetImage(s.rt)
	verhash := s.wp.GetVersionHash(s.rt)
	l := s.wp.LabelsForComponent("db-migrate")
	l["wordpress.presslabs.org/db-upgrade-for-hash"] = verhash

	out.Name = s.key.Name
	out.Namespace = s.key.Namespace
	out.Labels = l
	out.Annotations = map[string]string{
		"wordpress.presslabs.org/db-upgrade-for": image,
	}
	if err := controllerutil.SetControllerReference(s.wp, out, s.scheme); err != nil {
		return nil, err
	}

	out.Spec.BackoffLimit = &backoffLimit
	out.Spec.ActiveDeadlineSeconds = &activeDeadlineSeconds

	cmd := []string{"/bin/sh", "-c", "wp core update-db --network || wp core update-db && wp cache flush"}
	out.Spec.Template = *s.wp.JobPodTemplateSpec(s.rt, cmd...)

	out.Spec.Template.Labels = l

	return out, nil
}

func (s *dbUpgradeJobSyncer) GetErrorEventReason(err error) EventReason {
	if err != nil {
		return EventReasonDBUpgradeJobFailed
	}
	return EventReasonDBUpgradeJobUpdated
}
