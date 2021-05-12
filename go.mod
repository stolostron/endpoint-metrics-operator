module github.com/open-cluster-management/endpoint-metrics-operator

go 1.16

require (
	github.com/IBM/controller-filtered-cache v0.3.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/open-cluster-management/addon-framework v0.0.0-20210419013051-38730a847aff
	github.com/open-cluster-management/api v0.0.0-20210409125704-06f2aec1a73f
	github.com/open-cluster-management/multicluster-observability-operator v0.0.0-20210503035427-4955ac5d5746
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/openshift/client-go v0.0.0-20210331195552-cf6c2669e01f
	github.com/openshift/cluster-monitoring-operator v0.0.0-00010101000000-000000000000
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.47.1 // indirect
	github.com/prometheus/prometheus v1.8.2-0.20210217141258-a6be548dbc17 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kubectl v0.20.4
	sigs.k8s.io/controller-runtime v0.8.0
)

replace (
	github.com/go-logr/logr => github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.2.0
	github.com/jetstack/cert-manager => github.com/open-cluster-management/cert-manager v0.0.0-20200821135248-2fd523b053f5
	github.com/open-cluster-management/api => github.com/open-cluster-management/api v0.0.0-20201007180356-41d07eee4294
	github.com/openshift/api => github.com/openshift/api v0.0.0-20210331193751-3acddb19d360
	github.com/openshift/cluster-monitoring-operator => github.com/open-cluster-management/cluster-monitoring-operator v0.1.1-0.20210526090126-f3ac81b41e89
	golang.org/x/text => golang.org/x/text v0.3.5
	k8s.io/api => k8s.io/api v0.19.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.2
	k8s.io/client-go => k8s.io/client-go v0.19.2
)
