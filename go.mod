module github.com/stolostron/endpoint-metrics-operator

go 1.13

require (
	github.com/open-cluster-management/api v0.0.0-20201007180356-41d07eee4294
	github.com/open-cluster-management/multicluster-monitoring-operator v0.0.0-20201029062159-ac5203c2f91d
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/openshift/client-go v0.0.0-20201020082437-7737f16e53fc
	github.com/operator-framework/operator-sdk v0.18.0
	github.com/sykesm/zap-logfmt v0.0.4
	go.uber.org/zap v1.15.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kubectl v0.18.2
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	bitbucket.org/ww/goautoneg => github.com/markusthoemmes/goautoneg v0.0.0-20190713162725-c6008fefa5b1
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.25+incompatible
	github.com/go-logr/logr => github.com/go-logr/logr v0.2.1
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.2.0
	github.com/gorilla/websocket => github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/consul => github.com/hashicorp/consul v1.7.4
	github.com/jetstack/cert-manager => github.com/stolostron/cert-manager v0.0.0-20200821135248-2fd523b053f5
	github.com/mholt/caddy => github.com/caddyserver/caddy v1.0.5
	github.com/open-cluster-management/api => open-cluster-management.io/api v0.2.0
	github.com/openshift/api => github.com/openshift/api v0.0.0-20190924102528-32369d4db2ad // Required until https://github.com/operator-framework/operator-lifecycle-manager/pull/1241 is resolved
	github.com/openshift/origin => github.com/openshift/origin v1.2.0
	k8s.io/client-go => k8s.io/client-go v0.19.0
)
