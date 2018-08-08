/*
Copyright 2018 Pressinfra SRL.

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
	"context"
	"fmt"
	gosync "sync"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	"github.com/presslabs/wordpress-operator/pkg/controller/wordpress/sync"
)

var log = logf.Log.WithName(controllerName)

const controllerName = "wordpress-controller"

const (
	eventNormal  = "Normal"
	eventWarning = "Warning"
)

var rtMap struct {
	lock gosync.RWMutex
	m    map[types.NamespacedName]string
}

// Add creates a new Wordpress Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileWordpress{Client: mgr.GetClient(), scheme: mgr.GetScheme(), recorder: mgr.GetRecorder(controllerName)}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Wordpress
	err = c.Watch(&source.Kind{Type: &wordpressv1alpha1.Wordpress{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to WordpressRuntime
	err = c.Watch(&source.Kind{Type: &wordpressv1alpha1.WordpressRuntime{}}, &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(rt handler.MapObject) []reconcile.Request {
			rtMap.lock.RLock()
			defer rtMap.lock.RUnlock()
			var reconciles = []reconcile.Request{}
			for key, runtime := range rtMap.m {
				if runtime == rt.Meta.GetName() {
					reconciles = append(reconciles, reconcile.Request{NamespacedName: key})
				}
			}
			return reconciles
		}),
	})
	if err != nil {
		return err
	}

	// Watch for Deployment changes
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &wordpressv1alpha1.Wordpress{},
	})
	if err != nil {
		return err
	}

	// Watch for Service changes
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &wordpressv1alpha1.Wordpress{},
	})
	if err != nil {
		return err
	}

	// TODO(calind): watch for PVC, CronJobs, Jobs and Ingresses

	return nil
}

var _ reconcile.Reconciler = &ReconcileWordpress{}

// ReconcileWordpress reconciles a Wordpress object
type ReconcileWordpress struct {
	client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a Wordpress object and makes changes based on the state read
// and what is in the Wordpress.Spec
//
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wordpress.presslabs.org,resources=wordpresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wordpress.presslabs.org,resources=wordpressruntimes,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileWordpress) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Wordpress instance
	wp := &wordpressv1alpha1.Wordpress{}
	err := r.Get(context.TODO(), request.NamespacedName, wp)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Add Wordpress to the runtimes map
	rtMap.lock.Lock()
	rtMap.m[request.NamespacedName] = wp.Spec.Runtime
	rtMap.lock.Unlock()

	wp.SetDefaults()

	rt := &wordpressv1alpha1.WordpressRuntime{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: wp.Spec.Runtime}, rt)
	if err != nil {
		return reconcile.Result{}, err
	}
	rt.SetDefaults()

	syncers := []sync.Interface{
		sync.NewDeploymentSyncer(wp, rt, r.scheme),
		sync.NewServiceSyncer(wp, rt, r.scheme),
		sync.NewIngressSyncer(wp, rt, r.scheme),
		sync.NewWPCronSyncer(wp, rt, r.scheme),
		sync.NewDBUpgradeJobSyncer(wp, rt, r.scheme),
	}

	volSpec := rt.Spec.WebrootVolumeSpec
	if wp.Spec.WebrootVolumeSpec != nil {
		volSpec = wp.Spec.WebrootVolumeSpec
	}
	if volSpec.PersistentVolumeClaim != nil {
		syncers = append(syncers, sync.NewWebrootPVCSyncer(wp, rt, r.scheme))
	}

	volSpec = rt.Spec.MediaVolumeSpec
	if wp.Spec.MediaVolumeSpec != nil {
		volSpec = wp.Spec.MediaVolumeSpec
	}
	if volSpec != nil && volSpec.PersistentVolumeClaim != nil {
		syncers = append(syncers, sync.NewMediaPVCSyncer(wp, rt, r.scheme))
	}

	return reconcile.Result{}, r.sync(wp, syncers)
}

func (r *ReconcileWordpress) sync(wp *wordpressv1alpha1.Wordpress, syncers []sync.Interface) error {
	for _, s := range syncers {
		key := s.GetKey()
		existing := s.GetExistingObjectPlaceholder()

		op, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, key, existing, s.T)
		reason := string(s.GetErrorEventReason(err))

		log.Info(fmt.Sprintf("%T %s/%s %s", existing, key.Namespace, key.Name, op))

		if err != nil {
			r.recorder.Eventf(wp, eventWarning, reason, "%T %s/%s failed syncing: %s", existing, key.Namespace, key.Name, err)
			return err
		}
		if op != controllerutil.OperationNoop {
			r.recorder.Eventf(wp, eventNormal, reason, "%T %s/%s %s successfully", existing, key.Namespace, key.Name, op)
		}
	}
	return nil
}

func init() {
	rtMap.m = make(map[types.NamespacedName]string)
}
