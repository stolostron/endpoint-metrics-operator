// Copyright (c) 2021 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project.
package observabilityendpoint

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	"github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis"
	oav1beta1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/observability/v1beta1"
)

func getAllowlistCM() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metricsConfigMapName,
			Namespace: namespace,
		},
		Data: map[string]string{
			metricsConfigMapKey: `
names:
  - a
  - b
matches:
  - c
`},
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

func TestMetricsCollector(t *testing.T) {
	hubInfo := &HubInfo{
		ClusterName: "test-cluster",
		Endpoint:    "http://test-endpoint",
	}
	allowlistCM := getAllowlistCM()
	obsAddon := oav1beta1.ObservabilityAddonSpec{
		EnableMetrics: true,
		Interval:      60,
	}

	c := fake.NewFakeClient(allowlistCM)
	// Default deployment with instance count 1
	_, err := updateMetricsCollector(c, obsAddon, *hubInfo, testClusterID, 1, false)
	if err != nil {
		t.Fatalf("Failed to create metrics collector deployment: (%v)", err)
	}
	// Update deployment to reduce instance count to zero
	_, err = updateMetricsCollector(c, obsAddon, *hubInfo, testClusterID, 0, false)
	if err != nil {
		t.Fatalf("Failed to create metrics collector deployment: (%v)", err)
	}

	_, err = updateMetricsCollector(c, obsAddon, *hubInfo, testClusterID+"-update", 1, false)
	if err != nil {
		t.Fatalf("Failed to create metrics collector deployment: (%v)", err)
	}

	_, err = updateMetricsCollector(c, obsAddon, *hubInfo, testClusterID+"-update", 1, true)
	if err != nil {
		t.Fatalf("Failed to update metrics collector deployment: (%v)", err)
	}

	err = deleteMetricsCollector(c)
	if err != nil {
		t.Fatalf("Failed to delete metrics collector deployment: (%v)", err)
	}
}
