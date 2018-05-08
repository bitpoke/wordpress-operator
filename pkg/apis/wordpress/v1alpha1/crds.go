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

package v1alpha1

import (
	"fmt"

	kutil "github.com/appscode/kutil/apiextensions/v1beta1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	wpopenapi "github.com/presslabs/wordpress-operator/pkg/openapi"
)

const (
	wpapiPkg = "github.com/presslabs/wordpress-operator/pkg/apis/wordpress"
)

// Wordpress Custom Resource Definition
var (
	// ResourceWordpress contains the definition bits for Wordpress CRD
	ResourceWordpress = kutil.Config{
		Group:   SchemeGroupVersion.Group,
		Version: SchemeGroupVersion.Version,

		Kind:       ResourceKindWordpress,
		Plural:     "wordpresses",
		Singular:   "wordpress",
		ShortNames: []string{"wp"},

		SpecDefinitionName:    fmt.Sprintf("%s/%s.%s", wpapiPkg, SchemeGroupVersion.Version, ResourceKindWordpress),
		ResourceScope:         string(apiextensions.NamespaceScoped),
		GetOpenAPIDefinitions: wpopenapi.GetOpenAPIDefinitions,

		EnableValidation:        true,
		EnableStatusSubresource: true,
	}
	// ResourceWordpressCRDName is the fully qualified Wordpress CRD name (ie. worpresses.wordpress.presslabs.org)
	ResourceWordpressCRDName = fmt.Sprintf("%s.%s", ResourceWordpress.Plural, ResourceWordpress.Group)
	// ResourceWordpressCRD is the Custrom Resource Definition object for Wordpress
	ResourceWordpressCRD = kutil.NewCustomResourceDefinition(ResourceWordpress)
)

var CRDs = map[string]kutil.Config{
	ResourceWordpressCRDName: ResourceWordpress,
}
