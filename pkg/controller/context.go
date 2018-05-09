package controller

import (
	apiextenstions_clientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
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
	// KubeSharedInformerFactory can be used to obtain shared
	// SharedIndexInformer instances for Kubernetes types
	KubeSharedInformerFactory kubeinformers.SharedInformerFactory

	// Recorder to record events to
	Recorder record.EventRecorder

	// WordpressClient is a Presslabs Wordpress Operator clientset
	WordpressClient wpclientset.Interface
	// WordpressSharedInformerFactory can be used to obtain shared
	// SharedIndexInformer instances for Presslabs Wordpress Operator types
	WordpressSharedInformerFactory wpinformers.SharedInformerFactory

	// CRDClient is the clientset for Custom Resource Definitions
	CRDClient apiextenstions_clientset.ApiextensionsV1beta1Interface

	// InstallCRDs signals the controller whenever the install Worpdress CRDs
	InstallCRDs bool
	// RuntimeImage is the runtime image used for runnging Wordpress.
	RuntimeImage string
}
