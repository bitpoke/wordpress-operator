package syncer

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type objectSyncer struct {
	name   string
	owner  runtime.Object
	obj    runtime.Object
	syncFn controllerutil.MutateFn
	c      client.Client
	scheme *runtime.Scheme
}

func (s *objectSyncer) GetObject() interface{}   { return s.obj }
func (s *objectSyncer) GetOwner() runtime.Object { return s.owner }

func (s *objectSyncer) Sync(ctx context.Context) (SyncResult, error) {
	result := SyncResult{}

	key, err := getKey(s.obj)
	if err != nil {
		return result, err
	}

	result.Operation, err = controllerutil.CreateOrUpdate(ctx, s.c, s.obj, s.mutateFn())

	if err != nil {
		result.SetEventData(eventWarning, basicEventReason(s.name, err),
			fmt.Sprintf("%T %s failed syncing: %s", s.obj, key, err))
		log.Error(err, string(result.Operation), "key", key, "kind", fmt.Sprintf("%T", s.obj))
	} else {
		result.SetEventData(eventNormal, basicEventReason(s.name, err),
			fmt.Sprintf("%T %s %s successfully", s.obj, key, result.Operation))
		log.V(1).Info(string(result.Operation), "key", key, "kind", fmt.Sprintf("%T", s.obj))
	}

	return result, err
}

// Given an objectSyncer, returns a controllerutil.MutateFn which also sets the
// owner reference if the subject has one
func (s *objectSyncer) mutateFn() controllerutil.MutateFn {
	return func(existing runtime.Object) error {
		err := s.syncFn(existing)
		if err != nil {
			return err
		}
		if s.owner != nil {
			existingMeta, ok := existing.(metav1.Object)
			if !ok {
				return fmt.Errorf("%T is not a metav1.Object", existing)
			}
			ownerMeta, ok := s.owner.(metav1.Object)
			if !ok {
				return fmt.Errorf("%T is not a metav1.Object", s.owner)
			}
			err := controllerutil.SetControllerReference(ownerMeta, existingMeta, s.scheme)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// NewObjectSyncer creates a new kubernetes object syncer for a given object
// with an owner and persists data using controller-runtime's CreateOrUpdate.
// The name is used for logging and event emitting purposes and should be an
// valid go identifier in upper camel case. (eg. MysqlStatefulSet)
func NewObjectSyncer(name string, owner, obj runtime.Object, c client.Client, scheme *runtime.Scheme, syncFn controllerutil.MutateFn) Interface {
	return &objectSyncer{
		name:   name,
		owner:  owner,
		obj:    obj,
		c:      c,
		scheme: scheme,
		syncFn: syncFn,
	}
}
