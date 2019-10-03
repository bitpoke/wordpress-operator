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

	appsv1 "k8s.io/api/apps/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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
	"github.com/presslabs/wordpress-operator/pkg/internal/wordpress"
)

const controllerName = "wordpress-controller"

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

	subresources := []runtime.Object{
		&appsv1.Deployment{},
		&batchv1beta1.CronJob{},
		&corev1.PersistentVolumeClaim{},
		&corev1.Service{},
		&corev1.Secret{},
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
// +kubebuilder:rbac:groups=,resources=secrets;services;persistentvolumeclaims;events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=cronjobs;jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wordpress.presslabs.org,resources=wordpresses;wordpresses/status,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileWordpress) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Wordpress instance
	wp := wordpress.New(&wordpressv1alpha1.Wordpress{})
	err := r.Get(context.TODO(), request.NamespacedName, wp.Unwrap())
	if err != nil {
		return reconcile.Result{}, ignoreNotFound(err)
	}

	if updated, needsMigration := r.maybeMigrate(wp.Unwrap()); needsMigration {
		err = r.Update(context.TODO(), updated)
		return reconcile.Result{}, err
	}

	r.scheme.Default(wp.Unwrap())
	wp.SetDefaults()

	secretSyncer := sync.NewSecretSyncer(wp, r.Client, r.scheme)
	deploySyncer := sync.NewDeploymentSyncer(wp, secretSyncer.GetObject().(*corev1.Secret), r.Client, r.scheme)
	syncers := []syncer.Interface{
		secretSyncer,
		deploySyncer,
		sync.NewServiceSyncer(wp, r.Client, r.scheme),
		sync.NewIngressSyncer(wp, r.Client, r.scheme),
		sync.NewWPCronSyncer(wp, r.Client, r.scheme),
		// sync.NewDBUpgradeJobSyncer(wp, r.Client, r.scheme),
	}

	if wp.Spec.CodeVolumeSpec != nil && wp.Spec.CodeVolumeSpec.PersistentVolumeClaim != nil {
		syncers = append(syncers, sync.NewCodePVCSyncer(wp, r.Client, r.scheme))
	}

	if wp.Spec.MediaVolumeSpec != nil && wp.Spec.MediaVolumeSpec.PersistentVolumeClaim != nil {
		syncers = append(syncers, sync.NewMediaPVCSyncer(wp, r.Client, r.scheme))
	}

	if err = r.sync(syncers); err != nil {
		return reconcile.Result{}, err
	}

	oldStatus := wp.Status.DeepCopy()
	wp.Status.Replicas = deploySyncer.GetObject().(*appsv1.Deployment).Status.Replicas
	if oldStatus.Replicas != wp.Status.Replicas {
		if err := r.Status().Update(context.TODO(), wp.Unwrap()); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func ignoreNotFound(err error) error {
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}

func (r *ReconcileWordpress) sync(syncers []syncer.Interface) error {
	for _, s := range syncers {
		if err := syncer.Sync(context.TODO(), s, r.recorder); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileWordpress) maybeMigrate(wp *wordpressv1alpha1.Wordpress) (*wordpressv1alpha1.Wordpress, bool) {
	var needsMigration bool
	out := wp.DeepCopy()
	if len(out.Spec.Routes) == 0 {
		for i := range out.Spec.Domains {
			out.Spec.Routes = append(out.Spec.Routes, wordpressv1alpha1.RouteSpec{
				Domain: string(out.Spec.Domains[i]),
			})
			needsMigration = true
		}
	}
	out.Spec.Domains = nil
	return out, needsMigration
}
