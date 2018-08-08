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

package sync

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// EventReason is a type for storing kubernetes event reasons
type EventReason string

// Interface is the Wordpress syncer interface
type Interface interface {
	// GetKey returns the client.ObjectKey for looking up the dependant object
	GetKey() types.NamespacedName
	// GetExistingObjectPlaceholder returns a placeholder for existing object
	GetExistingObjectPlaceholder() runtime.Object
	// T transforms an objects to it's desired state
	T(in runtime.Object) (runtime.Object, error)
	// ErrorToEventReason returns an event reason for a T error
	GetErrorEventReason(err error) EventReason
}
