// Copyright (c) 2020 Red Hat, Inc.

package observabilityendpoint

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v2"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	oav1beta1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/observability/v1beta1"
)

const (
	hubInfoKey           = "hub-info.yaml"
	metricsCollectorName = "metrics-collector-deployment"
	selectorKey          = "component"
	selectorValue        = "metrics-collector"
	ocpPromURL           = "https://prometheus-k8s.openshift-monitoring.svc:9091"
	caMounthPath         = "/etc/serving-certs-ca-bundle"
	caVolName            = "serving-certs-ca-bundle"
	limitBytes           = 52428800
)

var (
	collectorImage         = os.Getenv("COLLECTOR_IMAGE")
	internalDefaultMetrics = []string{
		":node_memory_MemAvailable_bytes:sum",
		"cluster:capacity_cpu_cores:sum",
		"cluster:capacity_memory_bytes:sum",
		"cluster:container_cpu_usage:ratio",
		"cluster:container_spec_cpu_shares:ratio",
		"cluster:cpu_usage_cores:sum",
		"cluster:memory_usage:ratio",
		"cluster:memory_usage_bytes:sum",
		"cluster:usage:resources:sum",
		"cluster_infrastructure_provider",
		"cluster_version",
		"cluster_version_payload",
		"container_cpu_cfs_throttled_periods_total",
		"container_memory_cache",
		"container_memory_rss",
		"container_memory_swap",
		"container_memory_working_set_bytes",
		"container_network_receive_bytes_total",
		"container_network_receive_packets_dropped_total",
		"container_network_receive_packets_total",
		"container_network_transmit_bytes_total",
		"container_network_transmit_packets_dropped_total",
		"container_network_transmit_packets_total",
		"haproxy_backend_connections_total",
		"instance:node_cpu_utilisation:rate1m",
		"instance:node_load1_per_cpu:ratio",
		"instance:node_memory_utilisation:ratio",
		"instance:node_network_receive_bytes_excluding_lo:rate1m",
		"instance:node_network_receive_drop_excluding_lo:rate1m",
		"instance:node_network_transmit_bytes_excluding_lo:rate1m",
		"instance:node_network_transmit_drop_excluding_lo:rate1m",
		"instance:node_num_cpu:sum",
		"instance:node_vmstat_pgmajfault:rate1m",
		"instance_device:node_disk_io_time_seconds:rate1m",
		"instance_device:node_disk_io_time_weighted_seconds:rate1m",
		"kube_node_status_allocatable_cpu_cores",
		"kube_node_status_allocatable_memory_bytes",
		"kube_pod_container_resource_limits_cpu_cores",
		"kube_pod_container_resource_limits_memory_bytes",
		"kube_pod_container_resource_requests_cpu_cores",
		"kube_pod_container_resource_requests_memory_bytes",
		"kube_pod_info",
		"kube_resourcequota",
		"machine_cpu_cores",
		"machine_memory_bytes",
		"mixin_pod_workload",
		"node_cpu_seconds_total",
		"node_filesystem_avail_bytes",
		"node_filesystem_size_bytes",
		"node_memory_MemAvailable_bytes",
		"node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate",
		"node_namespace_pod_container:container_memory_cache",
		"node_namespace_pod_container:container_memory_rss",
		"node_namespace_pod_container:container_memory_swap",
		"node_namespace_pod_container:container_memory_working_set_bytes",
		"node_netstat_Tcp_OutSegs",
		"node_netstat_Tcp_RetransSegs",
		"node_netstat_TcpExt_TCPSynRetrans",
		"up",
	}
)

// HubInfo is the struct for hub info
type HubInfo struct {
	ClusterName string `yaml:"cluster-name"`
	Endpoint    string `yaml:"endpoint"`
}

func createDeployment(clusterName string, clusterID string, endpoint string,
	configs oav1beta1.ObservabilityAddonSpec, replicaCount int32) *appv1.Deployment {
	interval := configs.Interval
	commands := []string{
		"/usr/bin/telemeter-client",
		"--id=$(ID)",
		"--from=$(FROM)",
		"--to-upload=$(TO)",
		"--from-ca-file=" + caMounthPath + "/service-ca.crt",
		"--from-token-file=/var/run/secrets/kubernetes.io/serviceaccount/token",
		"--interval=" + fmt.Sprint(interval) + "s",
		"--label=\"cluster=" + clusterName + "\"",
		"--label=\"clusterID=" + clusterID + "\"",
		"--limit-bytes=" + strconv.Itoa(limitBytes),
	}
	for _, metrics := range internalDefaultMetrics {
		commands = append(commands, "--match={__name__=\""+metrics+"\"}")
	}
	return &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metricsCollectorName,
			Namespace: namespace,
			Annotations: map[string]string{
				ownerLabelKey: ownerLabelValue,
			},
		},
		Spec: appv1.DeploymentSpec{
			Replicas: int32Ptr(replicaCount),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					selectorKey: selectorValue,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						selectorKey: selectorValue,
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers: []v1.Container{
						{
							Name:    "metrics-collector",
							Image:   collectorImage,
							Command: commands,
							Env: []v1.EnvVar{
								{
									Name:  "FROM",
									Value: ocpPromURL,
								},
								{
									Name:  "TO",
									Value: endpoint,
								},
								{
									Name:  "ID",
									Value: clusterID,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      caVolName,
									MountPath: caMounthPath,
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: caVolName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: caConfigmapName,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func updateMetricsCollector(client kubernetes.Interface, hubInfo *v1.Secret,
	clusterID string, configs oav1beta1.ObservabilityAddonSpec, replicaCount int32) (bool, error) {
	hub := &HubInfo{}
	err := yaml.Unmarshal(hubInfo.Data[hubInfoKey], &hub)
	if err != nil {
		log.Error(err, "Failed to unmarshal hub info")
		return false, err
	}
	deployment := createDeployment(hub.ClusterName, clusterID, hub.Endpoint, configs, replicaCount)
	found, err := client.AppsV1().Deployments(namespace).Get(metricsCollectorName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = client.AppsV1().Deployments(namespace).Create(deployment)
			if err != nil {
				log.Error(err, "Failed to create metrics-collector deployment")
				return false, err
			}
			log.Info("Created metrics-collector deployment ")
		} else {
			log.Error(err, "Failed to get the metrics-collector deployment")
			return false, err
		}
	} else {
		if !reflect.DeepEqual(found.Spec, deployment.Spec) {
			deployment.ObjectMeta.ResourceVersion = found.ObjectMeta.ResourceVersion
			_, err = client.AppsV1().Deployments(namespace).Update(deployment)
			if err != nil {
				log.Error(err, "Failed to update metrics-collector deployment")
				return false, err
			}
			log.Info("Updated metrics-collector deployment ")
			return false, nil
		}
	}
	return true, nil
}

func deleteMetricsCollector(client kubernetes.Interface) error {
	_, err := client.AppsV1().Deployments(namespace).Get(metricsCollectorName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("The metrics collector deployment does not exist")
			return nil
		}
		log.Error(err, "Failed to check the metrics collector deployment")
		return err
	}
	// TODO: Should we set Replicas to zero instead?
	err = client.AppsV1().Deployments(namespace).Delete(metricsCollectorName, &metav1.DeleteOptions{})
	if err != nil {
		log.Error(err, "Failed to delete the metrics collector deployment")
		return err
	}
	log.Info("metrics collector deployment deleted")
	return nil
}

func int32Ptr(i int32) *int32 { return &i }
