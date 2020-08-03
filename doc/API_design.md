## ManagedCluster Monitoring API

### API Design:
The diagram is loctated in [here](https://swimlanes.io/u/sIBsY2gSF)

The requirement doc is located in [here](https://docs.google.com/document/d/1qawBUo8VcdBXuXzZl8sypIug1nLsUEm_5Yy0qENZ-aU)

EndpointObservability CR is namespace scoped and located in each cluster namespace in hub side if monitoring feature is enabled for that managed cluster. Hub operator will generate the default one in the cluster namespace and users can customize it later. One CR includes two sections: one for spec and the other for status.

**EndpointObservability** Spec: describe the specification and status for the metrics collector in one managed cluster

name | description | required | default | schema
---- | ----------- | -------- | ------- | ------
enableMetrics | Push metrics or not | yes | true | bool
metricsConfigs| Metrics collection configurations | yes | n/a | MetricsConfigs


**MetricsConfigs Spec**: describe the specification for metrics collected  from local prometheus and pushed to hub server

name | description | required | default | schema
---- | ----------- | -------- | ------- | ------
interval | Interval to collect&push metrics | yes | 1m | string


**EndpointObservability Status**: describe the status for current CR. It's updated by the metrics collector

name | description | required | default | schema
---- | ----------- | -------- | ------- | ------
metricsCollectionStatus | Collect/Push metrics successfully or not | no | true | bool
metricsCollectionError | Error encounted during metrics collect/push | no | n/a | string
logsCollectionStatus | Collect/Push logs successfully or not | no | true | bool
logsCollectionError | Error encounted during logs collect/push | no | n/a | string

### Samples

Here is a sample EndpointObservability CR

```
apiVersion: monitoring.open-cluster-management.io/v1alpha1
kind: EndpointObservability
metadata:
  name: sample-endpointmonitoring
spec:
  enableMetrics: true
  metricsConfigs:
    interval: 1m
```
