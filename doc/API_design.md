## ManagedCluster Monitoring API

### API Design:

The requirement doc is located in [here](https://docs.google.com/document/d/1qawBUo8VcdBXuXzZl8sypIug1nLsUEm_5Yy0qENZ-aU)

EndpointObservability CR is namespace scoped and located in each cluster namespace in hub side if monitoring feature is enabled for that managed cluster. Hub operator will generate the default one in the cluster namespace and users can customize it later. One CR includes two sections: one for spec and the other for status.

Group of this CR is observability.open-cluster-management.io, version is v1alpha1, kind is EndpointObservability

**EndpointObservability** Spec: describe the specification and status for the metrics collector in one managed cluster

name | description | required | default | schema
---- | ----------- | -------- | ------- | ------
enableMetrics | Push metrics or not | yes | true | bool
metricsConfigs| Metrics collection configurations | yes | n/a | MetricsConfigs


**MetricsConfigs Spec**: describe the specification for metrics collected  from local prometheus and pushed to hub server

name | description | required | default | schema
---- | ----------- | -------- | ------- | ------
interval | Interval for the metrics collector push metrics to  hub server| yes | 1m | string


**EndpointObservability Status**: describe the status for current CR. It's updated by the metrics collector

name | description | required | default | schema
---- | ----------- | -------- | ------- | ------
conditions | Conditions contains the different condition statuses for this managed cluster | no | [] | []Condtions

**Condition**: describe the condition status for current CR.

name | description | required | default | schema
---- | ----------- | -------- | ------- | ------
lastTransitionTime | Last time the condition transit from one status to another | yes | n/a | Time
status | Status of the condition, one of True, False, Unknown | yes | n/a | string
reason | (brief) reason for the condition's last transition | yes | n/a | string
message | Human readable message indicating details about last transition | yes | n/a | string
type | Type of node condition | yes | n/a | string



### Samples

Here is a sample EndpointObservability CR

```
apiVersion: observability.open-cluster-management.io/v1alpha1
kind: EndpointObservability
metadata:
  name: sample-endpointmonitoring
spec:
  enableMetrics: true
  metricsConfigs:
    interval: 1m
status:
  conditions:
    - type: Available
      status: 'True'
      lastTransitionTime: '2020-07-23T16:18:46Z'
      reason: ClientCreated
      message: The metrics collector client deployment created
```
