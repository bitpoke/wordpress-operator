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
	apiextenstions_util "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/golang/glog"
	apiextenstions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	wpapi "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	"github.com/presslabs/wordpress-operator/pkg/controller"
)

const (
	controllerName = "wordpress-controller"
	maxRetries     = 5
	threadiness    = 4
)

type Controller struct {
	*controller.Context
	*WordpressContext
}

func NewController(ctx *controller.Context) (c *Controller, err error) {
	c = &Controller{}
	c.Context = ctx

	c.initWordpressWorker()

	return
}

// Run starts the control loop for the Wordpress Controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	crds := []*apiextenstions.CustomResourceDefinition{
		wpapi.ResourceWordpressCRD,
	}

	if c.InstallCRDs {
		if err := c.installCRDs(crds); err != nil {
			glog.Fatalf(err.Error())
			return
		}
	}
	if err := c.waitForCRDs(crds); err != nil {
		glog.Fatalf(err.Error())
		return
	}

	glog.Infof("Starting %s control loops", controllerName)

	c.wpQueue.Run(stopCh)

	<-stopCh
	glog.Infof("Stopping %s control loops", controllerName)
}

func (c *Controller) installCRDs(crds []*apiextenstions.CustomResourceDefinition) error {
	glog.Info("Registering Custom Resource Definitions")

	if err := apiextenstions_util.RegisterCRDs(c.CRDClient, crds); err != nil {
		return err
	}

	return nil
}

func (c *Controller) waitForCRDs(crds []*apiextenstions.CustomResourceDefinition) error {
	glog.Info("Waiting for Custom Resource Definitions to become available")
	return apiextenstions_util.WaitForCRDReady(c.CRDClient.RESTClient(), crds)
}
