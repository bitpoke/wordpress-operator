package controller

import (
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"

	wpclientset "github.com/presslabs/wordpress-operator/pkg/client/clientset/versioned"
	wpinformers "github.com/presslabs/wordpress-operator/pkg/client/informers/externalversions"
)

// Context contains various types that are used by controller implementations.
// We purposely don't have specific informers/listers here, and instead keep
// a reference to a SharedInformerFactory so that controllers can choose
// themselves which listers are required.
type Context struct {
	// RESTConfig is the configuration for the REST client
	RESTConfig *rest.Config

	// KubeClient is a Kubernetes clientset
	KubeClient kubernetes.Interface
	// Recorder to record events to
	Recorder record.EventRecorder
	// KubeSharedInformerFactory can be used to obtain shared
	// SharedIndexInformer instances for Kubernetes types
	KubeSharedInformerFactory kubeinformers.SharedInformerFactory

	// WordpressClient is a Presslabs Wordpress Operator clientset
	WordpressClient wpclientset.Interface

	// WordpressSharedInformerFactory can be used to obtain shared
	// SharedIndexInformer instances for Presslabs Wordpress Operator types
	WordpressSharedInformerFactory wpinformers.SharedInformerFactory
}
