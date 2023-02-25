module github.com/bitpoke/wordpress-operator

go 1.16

require (
	github.com/appscode/mergo v0.3.6
	github.com/cooleo/slugify v0.0.0-20161029032441-81db6b52442d
	github.com/go-logr/logr v0.4.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/presslabs/controller-util v0.3.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.7.0

	// kubernetes
	k8s.io/api v0.21.4
	k8s.io/apimachinery v0.21.4
	k8s.io/client-go v0.21.4
	k8s.io/klog/v2 v2.10.0
	sigs.k8s.io/controller-runtime v0.9.7
)
