# endpoint-monitoring-operator

## Overview

The endpoint-monitoring-operator is a component of ACM observability feature. It is designed to install into Spoke Cluster.


## Developer Guide
The guide is used for developer to build and install the endpoint-monitoring-operator . It can be running in [kind][install_kind] if you don't have a OCP environment.

### Prerequisites

- [git][git_tool]
- [go][go_tool] version v1.13.9+.
- [docker][docker_tool] version 19.03+.
- [kubectl][kubectl_tool] version v1.14+.
- Access to a Kubernetes v1.11.3+ cluster.

### Install the Operator SDK CLI

Follow the steps in the [installation guide][install_guide] to learn how to install the Operator SDK CLI tool. It requires [version v0.17.0][operator_sdk_v0.17.0].
Or just use this command to download `operator-sdk` for Mac:
```
curl -L https://github.com/operator-framework/operator-sdk/releases/download/v0.17.0/operator-sdk-v0.17.0-x86_64-apple-darwin -o operator-sdk
```

### Build the Operator

- git clone this repository.
- `export GOPRIVATE=github.com/open-cluster-management`
- `go mod vendor`
- `operator-sdk build <repo>/<component>:<tag>` for example: quay.io/endpoint-monitoring-operator:v0.1.0.
- Update the image in `deploy/operator.yaml`.
- Update your namespace in `deploy/role_binding.yaml`
- Update the spec.global.serverUrl in `deploy/crds/monitoring.open-cluster-management.io_v1_endpointmonitoring_cr.yaml`. This is the observatorium api gateway url which exposed on hub cluster, such as observatorium-api-gateway-acm-monitoring.apps.calm-midge.dev05.red-chesterfield.com

### Deploy this Operator

1. Apply the manifests
```
kubectl apply -f deploy/crds/
kubectl apply -f deploy/

```
After installed successfully, you will see the following output:
`oc get pod`
```
NAME                                         READY   STATUS    RESTARTS   AGE
endpoint-monitoring-operator-68fbdbc66d-wm6rq   1/1     Running   0          46h
```
`oc get endpointmonitoring`
```
NAME                      AGE
example-endpointmonitoring   46h
```
**Notice**: To deploy the endpointmonitoring CR in local spoke cluster just for dev/test purpose. In real topology, the endpointmonitoring CR should be created in hub cluster, the endpoint-monitoring-operator should talk to api server of hub cluster to watch those CRs, and then perform changes on spoke cluster. 

2. The endpoint monitoring operator will create/update the configmap cluster-monitoring-config in openshift-monitoring namespace, based on the related info defined in the EndpointMoitoring CR. The changes will be applied automatically after several minutes. You can apply the changes immediately by invoking command below
```
oc scale --replicas=2 statefulset --all -n openshift-monitoring; oc scale --replicas=1 deployment --all -n openshift-monitoring
```

### View metrics in dashboard
Access Grafana console in hub cluster at https://{YOUR_DOMAIN}/grafana, view the metrics in the dashboard named "ACM:Managed Cluster Monitoring"


[install_kind]: https://github.com/kubernetes-sigs/kind
[install_guide]: https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md
[git_tool]:https://git-scm.com/downloads
[go_tool]:https://golang.org/dl/
[docker_tool]:https://docs.docker.com/install/
[kubectl_tool]:https://kubernetes.io/docs/tasks/tools/install-kubectl/
[operator_sdk_v0.17.0]:https://github.com/operator-framework/operator-sdk/releases/tag/v0.17.0
