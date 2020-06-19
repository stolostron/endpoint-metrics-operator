// Copyright (c) 2020 Red Hat, Inc.

package endpointmonitoring

import (
	"strings"
	"testing"

	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	configv1 "github.com/openshift/api/config/v1"
	fakeconfigclient "github.com/openshift/client-go/config/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	apiServerURL = "http://example.com"
	clusterID    = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
)

func TestCreateClusterMonitoringConfig(t *testing.T) {

	client := fake.NewFakeClientWithScheme(scheme.Scheme, []runtime.Object{}...)

	version := &configv1.ClusterVersion{
		ObjectMeta: metav1.ObjectMeta{Name: "version"},
		Spec: configv1.ClusterVersionSpec{
			ClusterID: configv1.ClusterID("xxxx-xxxx"),
		},
	}
	ocpClient := fakeconfigclient.NewSimpleClientset(version)

	labelConfigs := []monv1.RelabelConfig{
		{
			SourceLabels: []string{"__name__"},
			TargetLabel:  "cluster",
			Replacement:  labelValue,
		},
	}
	err := updateClusterMonitoringConfig(client, ocpClient, "apiServerURL", &labelConfigs)
	if err != nil {
		t.Errorf("Update configmap has error: %v", err)
	}

	cm, _ := getConfigMap(client)
	if cm.Data == nil {
		t.Errorf("Update configmap is failed")
	}
}

func TestUpdateClusterMonitoringConfig(t *testing.T) {

	cases := []struct {
		name    string
		rawData []byte
	}{
		{
			name: "update configmap with empty prometheusK8s",
			rawData: []byte(`
"http":
  "httpProxy": "test"
  "httpsProxy": "test"
`),
		},
		{
			name: "update configmap with empty remoteWrite",
			rawData: []byte(`
"http":
  "httpProxy": "test"
  "httpsProxy": "test"
prometheusK8s:
  externalLabels: null
  hostport: ""
  nodeSelector: null
  remoteWrite: null
`),
		},
		{
			name: "update configmap with the full values",
			rawData: []byte(`
"http":
  "httpProxy": "test"
  "httpsProxy": "test"
prometheusK8s:
  externalLabels: null
  hostport: ""
  nodeSelector: null
  remoteWrite:
  - url: http://observatorium/api/metrics/v1/write
    writeRelabelConfigs:
    - replacement: hub_cluster
      sourceLabels:
      - __name__
      targetLabel: cluster
    - replacement: xxxx-xxxx
      sourceLabels:
      - __name__
      targetLabel: cluster_id
  - url: http://apiServerURL/api/metrics/v1/write  
  - url: http://apiServerURL/api/metrics/v1/write
    writeRelabelConfigs:
    - replacement: spoke_cluster
      sourceLabels:
      - __name__
      targetLabel: cluster
    - replacement: xxxx-xxxx
      sourceLabels:
      - __name__
      targetLabel: cluster_id
  resources: null
  retention: ""
  tolerations: null
  volumeClaimTemplate: null
`),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cm := &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: cmNamespace,
				},
				Data: map[string]string{
					configKey: string(c.rawData),
				},
			}
			client := fake.NewFakeClientWithScheme(scheme.Scheme, cm)

			version := &configv1.ClusterVersion{
				ObjectMeta: metav1.ObjectMeta{Name: "version"},
				Spec: configv1.ClusterVersionSpec{
					ClusterID: configv1.ClusterID("xxxx-xxxx"),
				},
			}
			ocpClient := fakeconfigclient.NewSimpleClientset(version)

			labelConfigs := []monv1.RelabelConfig{
				{
					SourceLabels: []string{"__name__"},
					TargetLabel:  "cluster",
					Replacement:  labelValue,
				},
			}
			err := updateClusterMonitoringConfig(client, ocpClient, "apiServerURL", &labelConfigs)
			if err != nil {
				t.Errorf("Update configmap has error: %v", err)
			}

			configmap, _ := getConfigMap(client)
			if !strings.Contains(configmap.Data[configKey], "httpsProxy: test") {
				t.Errorf("Missed the original data in configmap %v", configmap.Data[configKey])
			}
		})
	}
}

func TestGetClusterIDSuccess(t *testing.T) {
	version := &configv1.ClusterVersion{
		ObjectMeta: metav1.ObjectMeta{Name: "version"},
		Spec: configv1.ClusterVersionSpec{
			ClusterID: configv1.ClusterID(clusterID),
		},
	}
	client := fakeconfigclient.NewSimpleClientset(version)
	tmpClusterID, _ := getClusterID(client)
	if tmpClusterID != clusterID {
		t.Errorf("OCP ClusterID (%v) is not the expected (%v)", tmpClusterID, clusterID)
	}
}

func TestGetClusterIDFailed(t *testing.T) {
	inf := &configv1.Infrastructure{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
		Status: configv1.InfrastructureStatus{
			APIServerURL: apiServerURL,
		},
	}
	client := fakeconfigclient.NewSimpleClientset(inf)
	_, err := getClusterID(client)
	if err == nil {
		t.Errorf("Should throw the error since there is no clusterversion defined")
	}
}
