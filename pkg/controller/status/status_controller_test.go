// Copyright (c) 2021 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project.
package status

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	"github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis"
	oav1beta1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/observability/v1beta1"
)

const (
	name            = "observability-addon"
	testNamespace   = "test-ns"
	testHubNamspace = "test-hub-ns"
)

func newObservabilityAddon(name string, ns string) *oav1beta1.ObservabilityAddon {
	return &oav1beta1.ObservabilityAddon{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: oav1beta1.ObservabilityAddonSpec{
			EnableMetrics: true,
			Interval:      60,
		},
	}
}

func init() {
	logf.SetLogger(logf.ZapLogger(true))

	s := scheme.Scheme
	addonv1alpha1.AddToScheme(s)
	apis.AddToScheme(s)

	namespace = testNamespace
	hubNamespace = testHubNamspace
}

func TestStatusController(t *testing.T) {

	hubClient := fake.NewFakeClient()
	c := fake.NewFakeClient()

	r := &ReconcileStatus{
		client:    c,
		hubClient: hubClient,
	}

	// test error in reconcile if missing obervabilityaddon
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "install",
			Namespace: testNamespace,
		},
	}
	_, err := r.Reconcile(req)
	if err == nil {
		t.Fatalf("reconcile: miss the error for missing obervabilityaddon")
	}

	// test status in local pushed to hub
	err = hubClient.Create(context.TODO(), newObservabilityAddon(name, testHubNamspace))
	if err != nil {
		t.Fatalf("failed to create hub oba to install: (%v)", err)
	}

	oba := newObservabilityAddon(name, testNamespace)
	oba.Status = oav1beta1.ObservabilityAddonStatus{
		Conditions: []oav1beta1.StatusCondition{
			{
				Type:    "Deployed",
				Status:  metav1.ConditionTrue,
				Reason:  "Deployed",
				Message: "Metrics collector deployed",
			},
		},
	}
	err = c.Create(context.TODO(), oba)
	if err != nil {
		t.Fatalf("failed to create oba to install: (%v)", err)
	}
	req = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "install",
			Namespace: testNamespace,
		},
	}
	_, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("Failed to reconcile: (%v)", err)
	}
	hubObsAddon := &oav1beta1.ObservabilityAddon{}
	err = hubClient.Get(context.TODO(), types.NamespacedName{Name: obAddonName, Namespace: testHubNamspace}, hubObsAddon)
	if err != nil {
		t.Fatalf("Failed to get oba in hub: (%v)", err)
	}

	if hubObsAddon.Status.Conditions == nil || len(hubObsAddon.Status.Conditions) != 1 {
		t.Fatalf("No correct status set in hub observabilityaddon: (%v)", hubObsAddon)
	} else if hubObsAddon.Status.Conditions[0].Type != "Deployed" {
		t.Fatalf("Wrong status type: (%v)", hubObsAddon.Status)
	}
}
