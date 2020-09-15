// Copyright (c) 2020 Red Hat, Inc.
package observabilityendpoint

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	oav1beta1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/observability/v1beta1"
)

var (
	hubInfo = &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hubConfigName,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			hubInfoKey: []byte(`
cluster-name: "test-cluster"
endpoint: "http://test-endpoint"
`),
		},
	}

	whitelistCM = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metricsConfigMapName,
			Namespace: namespace,
		},
		Data: map[string]string{metricsConfigMapKey: `
names:
	- a
	- b
`},
	}
)

func TestMetricsCollector(t *testing.T) {
	kubeClient := fake.NewSimpleClientset(whitelistCM)

	configs := &oav1beta1.ObservabilityAddonSpec{
		EnableMetrics: true,
		Interval:      60,
	}
	// Default deployment with instance count 1
	_, err := updateMetricsCollector(kubeClient, hubInfo, testClusterID, *configs, 1)
	if err != nil {
		t.Fatalf("Failed to create metrics collector deployment: (%v)", err)
	}
	// Update deployment to reduce instance count to zero
	_, err = updateMetricsCollector(kubeClient, hubInfo, testClusterID, *configs, 0)
	if err != nil {
		t.Fatalf("Failed to create metrics collector deployment: (%v)", err)
	}

	_, err = updateMetricsCollector(kubeClient, hubInfo, testClusterID+"-update", *configs, 1)
	if err != nil {
		t.Fatalf("Failed to create metrics collector deployment: (%v)", err)
	}

	err = deleteMetricsCollector(kubeClient)
	if err != nil {
		t.Fatalf("Failed to delete metrics collector deployment: (%v)", err)
	}
}
