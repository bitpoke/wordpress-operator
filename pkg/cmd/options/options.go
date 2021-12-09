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

import (
	"io/ioutil"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/spf13/pflag"
)

var (
	// GitCloneImage is the image used by the init container that clones the code.
	GitCloneImage = "docker.io/library/buildpack-deps:stretch-scm"

	// WordpressRuntimeImage is the base image used to run your code.
	WordpressRuntimeImage = "docker.io/bitpoke/wordpress-runtime:5.8.2"

	// IngressClass is the default ingress class used used for creating WordPress ingresses.
	IngressClass = ""

	// LeaderElection determines whether or not to use leader election when starting the manager.
	LeaderElection = false

	// LeaderElectionNamespace determines the namespace in which the leader election resource will be created.
	LeaderElectionNamespace = namespace()

	// LeaderElectionID determines the name of the resource that leader election will use for holding the leader lock.
	LeaderElectionID = "q7e8p7jr.wordpress-operator.bitpoke.io"

	// MetricsBindAddress is the TCP address that the controller should bind to for serving prometheus metrics.
	// It can be set to "0" to disable the metrics serving.
	MetricsBindAddress = ":8080"

	// HealthProbeBindAddress is the TCP address that the controller should bind to for serving health probes.
	HealthProbeBindAddress = ":8081"
)

func namespace() string {
	if ns := os.Getenv("KUBE_NAMESPACE"); ns != "" {
		return ns
	}

	if ns := os.Getenv("MY_POD_NAMESPACE"); ns != "" {
		return ns
	}

	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}

	return corev1.NamespaceDefault
}

// AddToFlagSet set command line arguments.
func AddToFlagSet(flag *pflag.FlagSet) {
	flag.StringVar(&GitCloneImage, "git-clone-image", GitCloneImage, "The image used when cloning code from git.")
	flag.StringVar(&WordpressRuntimeImage, "wordpress-runtime-image", WordpressRuntimeImage, "The base image used for Wordpress.")
	flag.StringVar(&IngressClass, "ingress-class", IngressClass, "The default ingress class for WordPress sites.")
	flag.BoolVar(&LeaderElection, "leader-election", LeaderElection, "Enables or disables controller leader election.")
	flag.StringVar(&LeaderElectionNamespace, "leader-election-namespace", LeaderElectionNamespace, "The namespace in which the leader election resource will be created.")
	flag.StringVar(&LeaderElectionID, "leader-election-id", LeaderElectionID, "The name of the resource that leader election will use for holding the leader lock.")
	flag.StringVar(&MetricsBindAddress, "metrics-addr", MetricsBindAddress, "The TCP address that the controller should bind to for serving prometheus metrics."+
		" It can be set to \"0\" to disable the metrics serving.")
	flag.StringVar(&HealthProbeBindAddress, "healthz-addr", HealthProbeBindAddress, "The TCP address that the controller should bind to for serving health probes.")
}
