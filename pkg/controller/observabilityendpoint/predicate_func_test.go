// Copyright (c) 2021 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project.
package observabilityendpoint

import (
	"testing"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestPredFunc(t *testing.T) {
	name := "test-obj"
	caseList := []struct {
		caseName       string
		namespace      string
		create         bool
		update         bool
		delete         bool
		expectedCreate bool
		expectedUpdate bool
		expectedDelete bool
	}{
		{
			caseName:       "All false",
			namespace:      testNamespace,
			create:         false,
			update:         false,
			delete:         false,
			expectedCreate: false,
			expectedUpdate: false,
			expectedDelete: false,
		},
		{
			caseName:       "All true",
			namespace:      testNamespace,
			create:         true,
			update:         true,
			delete:         true,
			expectedCreate: true,
			expectedUpdate: true,
			expectedDelete: true,
		},
		{
			caseName:       "All true for cluster scope obj",
			namespace:      "",
			create:         true,
			update:         true,
			delete:         true,
			expectedCreate: true,
			expectedUpdate: true,
			expectedDelete: true,
		},
	}

	for _, c := range caseList {
		t.Run(c.caseName, func(t *testing.T) {
			pred := getPred(name, c.namespace, c.create, c.update, c.delete)
			ce := event.CreateEvent{
				Meta: &metav1.ObjectMeta{
					Name:      name,
					Namespace: c.namespace,
				},
			}
			if c.expectedCreate {
				if !pred.CreateFunc(ce) {
					t.Fatalf("pre func return false on applied createevent in case: (%v)", c.caseName)
				}
				ce.Meta.SetName(name + "test")
				if pred.CreateFunc(ce) {
					t.Fatalf("pre func return true on different obj name in case: (%v)", c.caseName)
				}
			} else {
				if pred.CreateFunc(ce) {
					t.Fatalf("pre func return true on non-applied createevent in case: (%v)", c.caseName)
				}
			}

			ue := event.UpdateEvent{
				MetaOld: &metav1.ObjectMeta{
					Name:            name,
					Namespace:       c.namespace,
					ResourceVersion: "1",
				},
				MetaNew: &metav1.ObjectMeta{
					Name:            name,
					Namespace:       c.namespace,
					ResourceVersion: "2",
				},
				ObjectNew: &v1.Deployment{
					TypeMeta: metav1.TypeMeta{
						APIVersion: v1.SchemeGroupVersion.String(),
						Kind:       "Deployment",
					},
					Spec: v1.DeploymentSpec{
						Replicas: int32Ptr(2),
					},
				},
				ObjectOld: &v1.Deployment{
					TypeMeta: metav1.TypeMeta{
						APIVersion: v1.SchemeGroupVersion.String(),
						Kind:       "Deployment",
					},
					Spec: v1.DeploymentSpec{
						Replicas: int32Ptr(1),
					},
				},
			}
			if c.expectedUpdate {
				if !pred.UpdateFunc(ue) {
					t.Fatalf("pre func return false on applied updateevent in case: (%v)", c.caseName)
				}
				ue.MetaNew.SetResourceVersion("1")
				if pred.UpdateFunc(ue) {
					t.Fatalf("pre func return true on same resource version in case: (%v)", c.caseName)
				}
				ue.MetaNew.SetResourceVersion("2")
				ue.ObjectNew.(*v1.Deployment).Spec.Replicas = int32Ptr(1)
				if pred.UpdateFunc(ue) {
					t.Fatalf("pre func return true on same deployment spec in case: (%v)", c.caseName)
				}
			} else {
				if pred.UpdateFunc(ue) {
					t.Fatalf("pre func return true on non-applied updateevent in case: (%v)", c.caseName)
				}
			}

			de := event.DeleteEvent{
				Meta: &metav1.ObjectMeta{
					Name:      name,
					Namespace: c.namespace,
				},
			}
			if c.expectedDelete {
				if !pred.DeleteFunc(de) {
					t.Fatalf("pre func return false on applied deleteevent in case: (%v)", c.caseName)
				}
				de.Meta.SetName(name + "test")
				if pred.DeleteFunc(de) {
					t.Fatalf("pre func return true on different obj name in case: (%v)", c.caseName)
				}
			} else {
				if pred.DeleteFunc(de) {
					t.Fatalf("HubInpre funcfoPred return true on deleteevent in case: (%v)", c.caseName)
				}
			}
		})
	}
}
