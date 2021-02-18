// Copyright (c) 2020 Red Hat, Inc.

package observabilityendpoint

import (
	"fmt"
	"reflect"
	"strings"

	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	oav1beta1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/observability/v1beta1"
)

func getPred(name string, namespace string,
	create bool, update bool, delete bool) predicate.Funcs {
	createFunc := func(e event.CreateEvent) bool {
		return false
	}
	updateFunc := func(e event.UpdateEvent) bool {
		return false
	}
	deleteFunc := func(e event.DeleteEvent) bool {
		return false
	}
	if create {
		createFunc = func(e event.CreateEvent) bool {
			if e.Meta.GetName() == name && e.Meta.GetNamespace() == namespace {
				return true
			}
			return false
		}
	}
	if update {
		updateFunc = func(e event.UpdateEvent) bool {
			if e.MetaNew.GetName() == name && e.MetaNew.GetNamespace() == namespace &&
				e.MetaNew.GetResourceVersion() != e.MetaOld.GetResourceVersion() {
				// also check objectNew string in case Kind is empty
				if strings.HasPrefix(fmt.Sprint(e.ObjectNew), "&Deployment") ||
					e.ObjectNew.GetObjectKind().GroupVersionKind().Kind == "Deployment" {
					if !reflect.DeepEqual(e.ObjectNew.(*v1.Deployment).Spec, e.ObjectOld.(*v1.Deployment).Spec) {
						return true
					}
				} else if e.MetaNew.GetName() == obAddonName ||
					e.ObjectNew.GetObjectKind().GroupVersionKind().Kind == "ObservabilityAddon" {
					if !reflect.DeepEqual(e.ObjectNew.(*oav1beta1.ObservabilityAddon).Spec,
						e.ObjectOld.(*oav1beta1.ObservabilityAddon).Spec) {
						return true
					}
				} else {
					return true
				}
			}
			return false
		}
	}
	if delete {
		deleteFunc = func(e event.DeleteEvent) bool {
			if e.Meta.GetName() == name && e.Meta.GetNamespace() == namespace {
				return true
			}
			return false
		}
	}
	return predicate.Funcs{
		CreateFunc: createFunc,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
	}
}
