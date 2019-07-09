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

	// RcloneImage is the image used for writing and deleting media files, from cloud storage.
	RcloneImage = "quay.io/presslabs/rclone@sha256:4436a1e2d471236eafac605b24a66f5f18910b6f9cde505db065506208f73f96"

	// WordpressRuntimeImage is the base image used to run your code
	WordpressRuntimeImage = "quay.io/presslabs/wordpress-runtime"
	// WordpressRuntimeTag represents the tag used for WordpressRuntimeImage
	WordpressRuntimeTag = "5.2-7.3.4-r164"
)

// AddToFlagSet set command line arguments
func AddToFlagSet(flag *pflag.FlagSet) {
	flag.StringVar(&GitCloneImage, "git-clone-image", GitCloneImage, "The image used when cloning code from git.")
	flag.StringVar(&RcloneImage, "rclone-image", RcloneImage, "The image used for managing media files with Rclone.")
	flag.StringVar(&WordpressRuntimeImage, "wordpress-runtime-image", WordpressRuntimeImage, "The base image used for Wordpress.")
	flag.StringVar(&WordpressRuntimeTag, "wordpress-runtime-tag", WordpressRuntimeTag, "The base tag for Wordpress image.")
}
