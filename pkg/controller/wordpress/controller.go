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
	"github.com/appscode/kutil/tools/queue"
	"github.com/golang/glog"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	wpclientset "github.com/presslabs/wordpress-operator/pkg/client/clientset/versioned"
	wpinformers "github.com/presslabs/wordpress-operator/pkg/client/informers/externalversions"
	wplister "github.com/presslabs/wordpress-operator/pkg/client/listers/wordpress/v1alpha1"
	"github.com/presslabs/wordpress-operator/pkg/controller"
)

const (
	controllerName = "wordpress-controller"
	maxRetries     = 5
	threadiness    = 4
)

type Controller struct {
	kubeClient          kubernetes.Interface
	kubeInformerFactory informers.SharedInformerFactory

	wpClient          wpclientset.Interface
	wpInformerFactory wpinformers.SharedInformerFactory

	recorder record.EventRecorder

	// Wordpress CRD
	wpQueue    *queue.Worker
	wpInformer cache.SharedIndexInformer
	wpLister   wplister.WordpressLister
}

func NewController(
	ctx *controller.Context,
) (c *Controller, err error) {
	c = &Controller{
		kubeClient:          ctx.KubeClient,
		kubeInformerFactory: ctx.KubeSharedInformerFactory,

		wpClient:          ctx.WordpressClient,
		wpInformerFactory: ctx.WordpressSharedInformerFactory,

		recorder: ctx.Recorder,
	}

	c.initWordpressWorker()

	return
}

// Run starts the control loop for the Wordpress Controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	glog.Infof("Starting %s control loops", controllerName)

	c.wpQueue.Run(stopCh)

	<-stopCh
	glog.Infof("Stopping %s control loops", controllerName)
}
