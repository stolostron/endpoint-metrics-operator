// Copyright (c) 2020 Red Hat, Inc.
package observabilityendpoint

import (
	"fmt"
	"testing"

	"github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis"
	oav1beta1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/observability/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestReportStatus(t *testing.T) {
	oa := newObservabilityAddon()
	objs := []runtime.Object{oa}
	s := scheme.Scheme
	if err := apis.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add oav1beta1 scheme: (%v)", err)
	}

	expectedStatus := []oav1beta1.StatusCondition{
		{
			Type:    "NotSupported",
			Status:  metav1.ConditionTrue,
			Reason:  "NotSupported",
			Message: "No Prometheus service found in this cluster",
		},
		{
			Type:    "Ready",
			Status:  metav1.ConditionTrue,
			Reason:  "Deployed",
			Message: "Metrics collector deployed and functional",
		},
		{
			Type:    "Disabled",
			Status:  metav1.ConditionTrue,
			Reason:  "Disabled",
			Message: "enableMetrics is set to False",
		},
	}

	statusList := []string{"NotSupported", "Ready", "Disabled"}
	s.AddKnownTypes(oav1beta1.SchemeGroupVersion, oa)
	c := fake.NewFakeClient(objs...)
	for i := range statusList {
		reportStatus(c, oa, statusList[i])
		if oa.Status.Conditions[0].Message != expectedStatus[i].Message || oa.Status.Conditions[0].Reason != expectedStatus[i].Reason || oa.Status.Conditions[0].Status != expectedStatus[i].Status || oa.Status.Conditions[0].Type != expectedStatus[i].Type {
			t.Errorf("Error: Status not updated. Expected: %s, Actual: %s", expectedStatus[i], fmt.Sprintf("%+v\n", oa.Status.Conditions[0]))
		}
	}

}
