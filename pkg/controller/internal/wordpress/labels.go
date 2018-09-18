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

package wordpress

import (
	"k8s.io/apimachinery/pkg/labels"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
)

// DefaultLabels returns a general label set to apply to objects, relative to the
// Wordpress API object
func DefaultLabels(wp *wordpressv1alpha1.Wordpress) labels.Set {
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
func LabelsForTier(wp *wordpressv1alpha1.Wordpress, tier string) labels.Set {
	l := DefaultLabels(wp)
	l["app.kubernetes.io/tier"] = tier
	return l
}

// LabelsForComponent returns a label set object with component label filled in
func LabelsForComponent(wp *wordpressv1alpha1.Wordpress, component string) labels.Set {
	l := DefaultLabels(wp)
	l["app.kubernetes.io/component"] = component
	return l
}

// WebPodLabels returns the labels suitable Wordpress Web Pods
func WebPodLabels(wp *wordpressv1alpha1.Wordpress) labels.Set {
	l := LabelsForTier(wp, "front")
	l["app.kubernetes.io/component"] = "web"

	return l
}

// JobPodLabels returns the labels suitable Wordpress Web Pods
func JobPodLabels(wp *wordpressv1alpha1.Wordpress) labels.Set {
	return LabelsForComponent(wp, "wp-cli")
}
