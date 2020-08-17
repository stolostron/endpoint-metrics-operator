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
- push the image to the repo


### Deploy this Operator

1. Prerequisite
- Update the image in `deploy/operator.yaml`, to use the image built out in above step.
- Update the value of env COLLECTOR_IMAGE in `deploy/operator.yaml`, for example: quay.io/open-cluster-management/metrics-collector:2.1.0-PR6-1b7cdb7b33bd9baed230a367465ec7238204648a.
- Update your namespace in `deploy/role_binding.yaml`.
- Create the pull secret that used to pull the images, and set it in `deploy/service_account.yaml`
- Create the secret hub-info-secret.
```
kind: Secret
apiVersion: v1
metadata:
  name: hub-info-secret
data:
    hub-info.yaml: ***
type: Opaque
``` 
for the content of hub-info.yaml, it's base64-encoded yaml. The original yaml content is like below.
```
{
  "cluster-name": "my_cluster",
  "endpoint": "http://observatorium-api-open-cluster-management-observability.apps.stage3.demo.red-chesterfield.com/api/v1/receive"
}
```
**cluster-name** is the name for your cluster, you can set any non-empty string for it. **endpoint** is the observatorium api gateway url which exposed on hub cluster 

2. Apply the manifests
```
kubectl apply -f deploy/crds/
kubectl apply -f deploy/

```
After installed successfully, you will see the following output:
`oc get pod`
```
NAME                                         READY   STATUS    RESTARTS   AGE
endpoint-monitoring-operator-68fbdbc66d-wm6rq   1/1     Running   0          46h
metrics-collector-deployment-57d84fcf9b-tnsd4   1/1     Running   0          46h
```
`oc get observabilityaddon`
```
NAME                      AGE
observability-addon   46h
```
**Notice**: To deploy the observabilityaddon CR in local managed cluster just for dev/test purpose. In real topology, the observabilityaddon CR will be created in hub cluster, the endpoint-monitoring-operator should talk to api server of hub cluster to watch those CRs, and then perform changes on managed cluster. 

### View metrics in dashboard
Access Grafana console in hub cluster at https://{YOUR_DOMAIN}/grafana, view the metrics in the dashboard named "ACM:Managed Cluster Monitoring"


[install_kind]: https://github.com/kubernetes-sigs/kind
[install_guide]: https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md
[git_tool]:https://git-scm.com/downloads
[go_tool]:https://golang.org/dl/
[docker_tool]:https://docs.docker.com/install/
[kubectl_tool]:https://kubernetes.io/docs/tasks/tools/install-kubectl/
[operator_sdk_v0.17.0]:https://github.com/operator-framework/operator-sdk/releases/tag/v0.17.0
