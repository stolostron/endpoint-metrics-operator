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
	metricsConfigMapName = "observability-metrics-whitelist"
	metricsConfigMapKey  = "metrics_list.yaml"
	metricsCollectorName = "metrics-collector-deployment"
	selectorKey          = "component"
	selectorValue        = "metrics-collector"
	ocpPromURL           = "https://prometheus-k8s.openshift-monitoring.svc:9091"
	caMounthPath         = "/etc/serving-certs-ca-bundle"
	caVolName            = "serving-certs-ca-bundle"
	mtlsCertName         = "observability-managed-cluster-certs"
	limitBytes           = 52428800
	defaultInterval      = "60s"
)

var (
	collectorImage = os.Getenv("COLLECTOR_IMAGE")
)

type MetricsWhitelist struct {
	NameList  []string `yaml:"names"`
	MatchList []string `yaml:"matches"`
}

// HubInfo is the struct for hub info
type HubInfo struct {
	ClusterName string `yaml:"cluster-name"`
	Endpoint    string `yaml:"endpoint"`
}

func createDeployment(clusterName string, clusterID string, endpoint string,
	configs oav1beta1.ObservabilityAddonSpec, whitelist MetricsWhitelist, replicaCount int32) *appv1.Deployment {
	interval := fmt.Sprint(configs.Interval) + "s"
	if fmt.Sprint(configs.Interval) == "" {
		interval = defaultInterval
	}
	commands := []string{
		"/usr/bin/telemeter-client",
		"--id=$(ID)",
		"--from=$(FROM)",
		"--to-upload=$(TO)",
		"--from-ca-file=" + caMounthPath + "/service-ca.crt",
		"--from-token-file=/var/run/secrets/kubernetes.io/serviceaccount/token",
		"--interval=" + interval,
		"--label=\"cluster=" + clusterName + "\"",
		"--label=\"clusterID=" + clusterID + "\"",
		"--limit-bytes=" + strconv.Itoa(limitBytes),
	}
	for _, metrics := range whitelist.NameList {
		commands = append(commands, "--match={__name__=\""+metrics+"\"}")
	}
	for _, match := range whitelist.MatchList {
		commands = append(commands, "--match={"+match+"}")
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
								{
									Name:      mtlsCertName,
									MountPath: "/tlscerts",
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
						{
							Name: mtlsCertName,
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: mtlsCertName,
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
	list := getMetricsWhitelist(client)
	deployment := createDeployment(hub.ClusterName, clusterID, hub.Endpoint, configs, list, replicaCount)
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

func getMetricsWhitelist(client kubernetes.Interface) MetricsWhitelist {
	l := &MetricsWhitelist{}
	cm, err := client.CoreV1().ConfigMaps(namespace).Get(metricsConfigMapName, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Failed to get configmap")
	} else {
		if cm.Data != nil {
			err = yaml.Unmarshal([]byte(cm.Data[metricsConfigMapKey]), l)
			if err != nil {
				log.Error(err, "Failed to unmarshal data in configmap")
			}
		}
	}
	return *l
}
