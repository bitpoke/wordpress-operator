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
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/golang/glog"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"

	wpapi "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	"github.com/presslabs/wordpress-operator/pkg/factory/wordpress"
)

const (
	dbMigrateJobName = "%s-db-migrate"
)

func (c *Controller) syncDBMigrate(wp *wpapi.Wordpress) error {
	glog.Infof("Syncing db migration job for %s/%s", wp.ObjectMeta.Namespace, wp.ObjectMeta.Name)

	wpf := wordpress.Generator{WP: wp}

	labels := wpf.Labels()
	labels["app.kubernetes.io/component"] = "db-migrate"

	var images []string

	for _, c := range wp.Spec.WebPodTemplate.Spec.Containers {
		images = append(images, c.Image)
	}

	image := strings.Join(images, ",")

	meta := c.objectMeta(wp, dbMigrateJobName)
	meta.Name = fmt.Sprintf("%s-%x", meta.Name, md5.Sum([]byte(image)))
	meta.Labels = labels
	meta.Annotations = map[string]string{
		"wordpress.presslabs.org/db-upgrade-for": image,
	}

	// After release of kubernetes 1.10.5 increase backoff limit
	var backoffLimit int32 = 0
	var activeDeadlineSeconds int64 = 86400

	cmd := []string{"/bin/sh", "-c", "wp core update-db --network || wp core update-db && wp cache flush"}
	podTemplate := *wpf.JobPodTemplateSpec(&corev1.PodTemplateSpec{}, cmd...)
	podTemplate.Labels = labels

	_, err := c.KubeClient.BatchV1().Jobs(meta.Namespace).Create(&batchv1.Job{
		ObjectMeta: c.ensureControllerReference(meta, wp),
		Spec: batchv1.JobSpec{
			BackoffLimit:          &backoffLimit,
			ActiveDeadlineSeconds: &activeDeadlineSeconds,
			Template:              podTemplate,
		},
	})

	if err == nil {
		glog.V(3).Infof("Created Job %s/%s.", meta.Namespace, meta.Name)
	} else if kerr.IsAlreadyExists(err) {
		glog.V(3).Infof("Job %s/%s already exists.", meta.Namespace, meta.Name)
		return nil
	}

	return err
}
