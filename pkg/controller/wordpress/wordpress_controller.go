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
	gosync "sync"

	appsv1 "k8s.io/api/apps/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/presslabs/controller-util/syncer"

	wordpressv1alpha1 "github.com/presslabs/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	"github.com/presslabs/wordpress-operator/pkg/controller/wordpress/internal/sync"
)

const controllerName = "wordpress-controller"

var rtMap = &gosync.Map{}

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
			var reconciles = []reconcile.Request{}
			rtMap.Range(func(key, runtime interface{}) bool {
				if runtime.(string) == rt.Meta.GetName() {
					reconciles = append(reconciles, reconcile.Request{NamespacedName: key.(types.NamespacedName)})
				}
				return true
			})
			return reconciles
		}),
	})
	if err != nil {
		return err
	}

	subresources := []runtime.Object{
		&appsv1.Deployment{},
		&batchv1beta1.CronJob{},
		&corev1.PersistentVolumeClaim{},
		&corev1.Service{},
		&extv1beta1.Ingress{},
	}

	for _, subresource := range subresources {
		err = c.Watch(&source.Kind{Type: subresource}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &wordpressv1alpha1.Wordpress{},
		})
		if err != nil {
			return err
		}
	}

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
// +kubebuilder:rbac:groups=,resources=services;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=cronjobs;jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
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
	rtMap.Store(request.NamespacedName, wp.Spec.Runtime)

	r.scheme.Default(wp)

	rt := &wordpressv1alpha1.WordpressRuntime{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: wp.Spec.Runtime}, rt)
	if err != nil {
		return reconcile.Result{}, err
	}
	r.scheme.Default(rt)

	syncers := []syncer.Interface{
		sync.NewDeploymentSyncer(wp, rt, r.Client, r.scheme),
		sync.NewServiceSyncer(wp, rt, r.Client, r.scheme),
		sync.NewIngressSyncer(wp, rt, r.Client, r.scheme),
		sync.NewWPCronSyncer(wp, rt, r.Client, r.scheme),
		sync.NewDBUpgradeJobSyncer(wp, rt, r.Client, r.scheme),
	}

	volSpec := rt.Spec.WebrootVolumeSpec
	if wp.Spec.WebrootVolumeSpec != nil {
		volSpec = wp.Spec.WebrootVolumeSpec
	}
	if volSpec.PersistentVolumeClaim != nil {
		syncers = append(syncers, sync.NewWebrootPVCSyncer(wp, rt, r.Client, r.scheme))
	}

	volSpec = rt.Spec.MediaVolumeSpec
	if wp.Spec.MediaVolumeSpec != nil {
		volSpec = wp.Spec.MediaVolumeSpec
	}
	if volSpec != nil && volSpec.PersistentVolumeClaim != nil {
		syncers = append(syncers, sync.NewMediaPVCSyncer(wp, rt, r.Client, r.scheme))
	}

	return reconcile.Result{}, r.sync(syncers)
}

func (r *ReconcileWordpress) sync(syncers []syncer.Interface) error {
	for _, s := range syncers {
		if err := syncer.Sync(context.TODO(), s, r.recorder); err != nil {
			return err
		}
	}
	return nil
}
