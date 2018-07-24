/*
Copyright 2018 The Kubernetes Authors.

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

package controllerutil

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// AlreadyOwnedError is an error returned if the object you are trying to assign
// a controller reference is already owned by another controller Object is the
// subject and Owner is the reference for the current owner
type AlreadyOwnedError struct {
	Object v1.Object
	Owner  v1.OwnerReference
}

func (e *AlreadyOwnedError) Error() string {
	return fmt.Sprintf("Object %s/%s is already owned by another %s controller %s", e.Object.GetNamespace(), e.Object.GetName(), e.Owner.Kind, e.Owner.Name)
}

func newAlreadyOwnedError(Object v1.Object, Owner v1.OwnerReference) *AlreadyOwnedError {
	return &AlreadyOwnedError{
		Object: Object,
		Owner:  Owner,
	}
}

// SetControllerReference sets owner as a Controller OwnerReference on owned.
// This is used for garbage collection of the owned object and for
// reconciling the owner object on changes to owned (with a Watch + EnqueueRequestForOwner).
// Since only one OwnerReference can be a controller, it returns an error if
// there is another OwnerReference with Controller flag set.
func SetControllerReference(owner, object v1.Object, scheme *runtime.Scheme) error {
	ro, ok := owner.(runtime.Object)
	if !ok {
		return fmt.Errorf("is not a %T a runtime.Object, cannot call SetControllerReference", owner)
	}

	gvk, err := apiutil.GVKForObject(ro, scheme)
	if err != nil {
		return err
	}

	// Create a new ref
	ref := *v1.NewControllerRef(owner, schema.GroupVersionKind{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind})

	existingRefs := object.GetOwnerReferences()
	fi := -1
	for i, r := range existingRefs {
		if referSameObject(ref, r) {
			fi = i
		} else if r.Controller != nil && *r.Controller {
			return newAlreadyOwnedError(object, r)
		}
	}
	if fi == -1 {
		existingRefs = append(existingRefs, ref)
	} else {
		existingRefs[fi] = ref
	}

	// Update owner references
	object.SetOwnerReferences(existingRefs)
	return nil
}

// Returns true if a and b point to the same object
func referSameObject(a, b v1.OwnerReference) bool {
	aGV, err := schema.ParseGroupVersion(a.APIVersion)
	if err != nil {
		return false
	}

	bGV, err := schema.ParseGroupVersion(b.APIVersion)
	if err != nil {
		return false
	}

	return aGV == bGV && a.Kind == b.Kind && a.Name == b.Name
}

// OperationType is the action result of a CreateOrUpdate call
type OperationType string

const ( // They should complete the sentence "Deployment default/foo has been ..."
	// OperationNoop means that the resource has not been changed
	OperationNoop = "unchanged"
	// OperationCreated means that a new resource has been created
	OperationCreated = "created"
	// OperationUpdated means that an existing resource has been updated
	OperationUpdated = "updated"
)

// CreateOrUpdate creates or updates a kubernetes resource. It takes in a key and
// a placeholder for the existing object and returns the operation executed
func CreateOrUpdate(ctx context.Context, c client.Client, key client.ObjectKey, existing runtime.Object, t TransformFn) (OperationType, error) {
	err := c.Get(ctx, key, existing)
	var obj runtime.Object

	if errors.IsNotFound(err) {
		// Create a new zero value object so that the in parameter of
		// TransformFn is always a "clean" object, with only Name and Namespace
		// set
		zero := reflect.New(reflect.TypeOf(existing).Elem()).Interface()

		// Set Namespace and Name from the lookup key
		zmeta, ok := zero.(v1.Object)
		if !ok {
			return OperationNoop, fmt.Errorf("is not a %T a metav1.Object, cannot call CreateOrUpdate", zero)
		}
		zmeta.SetNamespace(key.Namespace)
		zmeta.SetName(key.Name)

		// Apply the TransformFn
		obj, err = t(zero.(runtime.Object))
		if err != nil {
			return OperationNoop, err
		}

		// Create the new object
		err = c.Create(ctx, obj)
		if err != nil {
			return OperationNoop, err
		}

		return OperationCreated, err
	} else if err != nil {
		return OperationNoop, err
	} else {
		obj, err = t(existing.DeepCopyObject())
		if err != nil {
			return OperationNoop, err
		}

		if !reflect.DeepEqual(existing, obj) {
			err = c.Update(ctx, obj)
			if err != nil {
				return OperationNoop, err
			}

			return OperationUpdated, err
		}

		return OperationNoop, nil
	}
}

// TransformFn is a function which take in a kubernetes object and returns the
// desired state of that object.
// It is safe to mutate the object inside this function, since it's always
// called with an object's deep copy.
type TransformFn func(in runtime.Object) (runtime.Object, error)
