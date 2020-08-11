package observabilityendpoint

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
)

func TestMetricsCollector(t *testing.T) {
	kubeClient := fake.NewSimpleClientset([]runtime.Object{}...)

	configs := &oav1beta1.MetricsConfigsSpec{
		Interval: "1m",
	}
	err := createMetricsCollector(kubeClient, hubInfo, testClusterID, *configs)
	if err != nil {
		t.Fatalf("Failed to create metrics collector deployment: (%v)", err)
	}
	err = createMetricsCollector(kubeClient, hubInfo, testClusterID+"-update", *configs)
	if err != nil {
		t.Fatalf("Failed to create metrics collector deployment: (%v)", err)
	}

	err = deleteMetricsCollector(kubeClient)
	if err != nil {
		t.Fatalf("Failed to delete metrics collector deployment: (%v)", err)
	}
}
