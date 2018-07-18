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
	"reflect"

	"github.com/appscode/kutil/tools/queue"
	"github.com/golang/glog"
	"k8s.io/client-go/tools/cache"

	wpapi "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	// wpinformer "github.com/presslabs/wordpress-operator/pkg/client/informers/externalversions"
	wplister "github.com/presslabs/wordpress-operator/pkg/client/listers/wordpress/v1alpha1"
)

type WordpressContext struct {
	// Wordpress CRD
	wpQueue    *queue.Worker
	wpInformer cache.SharedIndexInformer
	wpLister   wplister.WordpressLister
}

func (c *Controller) initWordpressWorker() {
	inf := c.WordpressSharedInformerFactory.Wordpress().V1alpha1().Wordpresses()
	c.WordpressContext = &WordpressContext{
		wpInformer: inf.Informer(),
		wpLister:   inf.Lister(),
		wpQueue:    queue.New("wordpress", maxRetries, threadiness, c.reconcileWordpress),
	}

	c.wpInformer.AddEventHandler(queue.NewEventHandler(c.wpQueue.GetQueue(), func(old interface{}, new interface{}) bool {
		oldSpec, ok := old.(*wpapi.Wordpress)
		if !ok {
			return false
		}
		newSpec, ok := new.(*wpapi.Wordpress)
		if !ok {
			return false
		}
		return !reflect.DeepEqual(oldSpec.Spec, newSpec.Spec)
	}))
}

func (c *Controller) reconcileWordpress(key string) error {
	obj, exists, err := c.wpInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}
	if exists {
		glog.Infof("Sync/Add/Update for Wordpress %s", key)
		wp := obj.(*wpapi.Wordpress).DeepCopy().WithDefaults()

		if err := c.syncService(wp); err != nil {
			return err
		}

		if err := c.syncIngress(wp); err != nil {
			return err
		}

		if err := c.syncPVC(wp); err != nil {
			return err
		}

		if err := c.syncDeployment(wp); err != nil {
			return err
		}

		// if err := c.syncCron(wp); err != nil {
		// 	return err
		// }

		// if err := c.syncDBMigrate(wp); err != nil {
		// 	return err
		// }
	}
	return nil
}
