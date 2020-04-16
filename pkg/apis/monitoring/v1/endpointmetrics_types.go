package v1

import (
	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EndpointMetricsSpec defines the desired state of EndpointMetrics
type EndpointMetricsSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	GlobalConfig         GlobalConfigSpec       `json:"global"`
	MetricsCollectorList []MetricsCollectorSpec `json:"metricsCollectors"`
}

type GlobalConfigSpec struct {
	SeverURL       string            `json:"serverUrl"`
	TLSConfig      *monv1.TLSConfig  `json:"tlsConfig,omitempty"`
	ExternalLabels map[string]string `json:"externalLabels,omitempty"`
}

type MetricsCollectorSpec struct {
	Enable bool   `json:"enable"`
	Type   string `json:"type"`
}

// EndpointMetricsStatus defines the observed state of EndpointMetrics
type EndpointMetricsStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EndpointMetrics is the Schema for the endpointmetrics API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=endpointmetrics,scope=Namespaced
type EndpointMetrics struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EndpointMetricsSpec   `json:"spec,omitempty"`
	Status EndpointMetricsStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EndpointMetricsList contains a list of EndpointMetrics
type EndpointMetricsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EndpointMetrics `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EndpointMetrics{}, &EndpointMetricsList{})
}
