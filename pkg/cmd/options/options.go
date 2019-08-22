/*
Copyright 2019 Pressinfra SRL.

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

package options

import "github.com/spf13/pflag"

var (
	// GitCloneImage is the image used by the init container that clones the code.
	GitCloneImage = "docker.io/library/buildpack-deps:stretch-scm"

	// WordpressRuntimeImage is the base image used to run your code
	WordpressRuntimeImage = "quay.io/presslabs/wordpress-runtime"
	// WordpressRuntimeTag represents the tag used for WordpressRuntimeImage
	WordpressRuntimeTag = "5.2.2"

	// IngressClass is the default ingress class used used for creating WordPress ingresses
	IngressClass = ""
)

// AddToFlagSet set command line arguments
func AddToFlagSet(flag *pflag.FlagSet) {
	flag.StringVar(&GitCloneImage, "git-clone-image", GitCloneImage, "The image used when cloning code from git.")
	flag.StringVar(&WordpressRuntimeImage, "wordpress-runtime-image", WordpressRuntimeImage, "The base image used for Wordpress.")
	flag.StringVar(&WordpressRuntimeTag, "wordpress-runtime-tag", WordpressRuntimeTag, "The base tag for Wordpress image.")
	flag.StringVar(&IngressClass, "ingress-class", IngressClass, "The default ingress class for WordPress sites.")
}
