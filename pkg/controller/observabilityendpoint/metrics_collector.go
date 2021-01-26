// Copyright (c) 2021 Red Hat, Inc.

package observabilityendpoint

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	oav1beta1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/observability/v1beta1"
)

const (
	hubInfoKey           = "hub-info.yaml"
	metricsConfigMapName = "observability-metrics-whitelist"
	metricsConfigMapKey  = "metrics_list.yaml"
	metricsCollectorName = "metrics-collector-deployment"
	selectorKey          = "component"
	selectorValue        = "metrics-collector"
	caMounthPath         = "/etc/serving-certs-ca-bundle"
	caVolName            = "serving-certs-ca-bundle"
	mtlsCertName         = "observability-managed-cluster-certs"
	limitBytes           = 1073741824
	defaultInterval      = "60s"
)

const (
	kindClusterID   = "kind-cluster-id"
	kindClusterHost = "observatorium.hub"
	kindClusterIP   = "172.17.0.2"
)

var (
	collectorImage = os.Getenv("COLLECTOR_IMAGE")
	ocpPromURL     = "https://prometheus-k8s.openshift-monitoring.svc:9091"
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

func createDeployment(clusterID string, obsAddonSpec oav1beta1.ObservabilityAddonSpec,
	hubInfo HubInfo, whitelist MetricsWhitelist, replicaCount int32) *v1.Deployment {
	interval := fmt.Sprint(obsAddonSpec.Interval) + "s"
	if fmt.Sprint(obsAddonSpec.Interval) == "" {
		interval = defaultInterval
	}

	volumes := []corev1.Volume{
		{
			Name: mtlsCertName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: mtlsCertName,
				},
			},
		},
	}
	mounts := []corev1.VolumeMount{
		{
			Name:      mtlsCertName,
			MountPath: "/tlscerts",
		},
	}
	caFile := caMounthPath + "/service-ca.crt"
	if clusterID == "" {
		clusterID = hubInfo.ClusterName
		// deprecated ca bundle, only used for ocp 3.11 env
		caFile = "//run/secrets/kubernetes.io/serviceaccount/service-ca.crt"
	} else {
		volumes = append(volumes, corev1.Volume{
			Name: caVolName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: caConfigmapName,
					},
				},
			},
		})
		mounts = append(mounts, corev1.VolumeMount{
			Name:      caVolName,
			MountPath: caMounthPath,
		})
	}

	hostAlias := []corev1.HostAlias{}
	// patch for e2e test using kind cluster
	if clusterID == kindClusterID {
		ocpPromURL = "http://prometheus-k8s.openshift-monitoring.svc:9090"
		hostAlias = append(hostAlias, corev1.HostAlias{
			IP:        kindClusterIP,
			Hostnames: []string{kindClusterHost},
		})
	}

	commands := []string{
		"/usr/bin/telemeter-client",
		"--id=$(ID)",
		"--from=$(FROM)",
		"--to-upload=$(TO)",
		"--from-ca-file=" + caFile,
		"--from-token-file=/var/run/secrets/kubernetes.io/serviceaccount/token",
		"--interval=" + interval,
		"--label=\"cluster=" + hubInfo.ClusterName + "\"",
		"--label=\"clusterID=" + clusterID + "\"",
		"--limit-bytes=" + strconv.Itoa(limitBytes),
	}
	for _, metrics := range whitelist.NameList {
		commands = append(commands, "--match={__name__=\""+metrics+"\"}")
	}
	for _, match := range whitelist.MatchList {
		commands = append(commands, "--match={"+match+"}")
	}
	return &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metricsCollectorName,
			Namespace: namespace,
			Annotations: map[string]string{
				ownerLabelKey: ownerLabelValue,
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: int32Ptr(replicaCount),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					selectorKey: selectorValue,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						selectorKey: selectorValue,
					},
				},
				Spec: corev1.PodSpec{
					HostAliases:        hostAlias,
					ServiceAccountName: serviceAccountName,
					Containers: []corev1.Container{
						{
							Name:    "metrics-collector",
							Image:   collectorImage,
							Command: commands,
							Env: []corev1.EnvVar{
								{
									Name:  "FROM",
									Value: ocpPromURL,
								},
								{
									Name:  "TO",
									Value: hubInfo.Endpoint,
								},
								{
									Name:  "ID",
									Value: clusterID,
								},
							},
							VolumeMounts: mounts,
						},
					},
					Volumes: volumes,
				},
			},
		},
	}
}

func updateMetricsCollector(c client.Client, obsAddonSpec oav1beta1.ObservabilityAddonSpec,
	hubInfo HubInfo, clusterID string,
	replicaCount int32, forceRestart bool) (bool, error) {

	list := getMetricsWhitelist(c)
	deployment := createDeployment(clusterID, obsAddonSpec, hubInfo, list, replicaCount)
	found := &v1.Deployment{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: metricsCollectorName,
		Namespace: namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			err = c.Create(context.TODO(), deployment)
			if err != nil {
				log.Error(err, "Failed to create metrics-collector deployment")
				return false, err
			}
			log.Info("Created metrics-collector deployment ")
		} else {
			log.Error(err, "Failed to check the metrics-collector deployment")
			return false, err
		}
	} else {
		if !reflect.DeepEqual(found.Spec, deployment.Spec) {
			deployment.ObjectMeta.ResourceVersion = found.ObjectMeta.ResourceVersion
			err = c.Update(context.TODO(), deployment)
			if err != nil {
				log.Error(err, "Failed to update metrics-collector deployment")
				return false, err
			}
			log.Info("Updated metrics-collector deployment ")
		}
		if forceRestart {
			err := deletePod(c)
			if err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

func deleteMetricsCollector(client client.Client) error {
	found := &v1.Deployment{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: metricsCollectorName,
		Namespace: namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("The metrics collector deployment does not exist")
			return nil
		}
		log.Error(err, "Failed to check the metrics collector deployment")
		return err
	}
	err = client.Delete(context.TODO(), found)
	if err != nil {
		log.Error(err, "Failed to delete the metrics collector deployment")
		return err
	}
	log.Info("metrics collector deployment deleted")
	return nil
}

func int32Ptr(i int32) *int32 { return &i }

func getMetricsWhitelist(client client.Client) MetricsWhitelist {
	l := &MetricsWhitelist{}
	cm := &corev1.ConfigMap{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: metricsConfigMapName,
		Namespace: namespace}, cm)
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

func deletePod(c client.Client) error {
	podList := &corev1.PodList{}
	options := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{
			selectorKey: selectorValue,
		}),
	}
	err := c.List(context.TODO(), podList, options...)
	if err != nil && errors.IsNotFound(err) {
		log.Error(err, "Failed to list pods of metrics collector")
		return err
	}
	for index := range podList.Items {
		pod := podList.Items[index]
		err := c.Delete(context.TODO(), &pod)
		if err != nil {
			log.Error(err, "Failed to delete pod", "name", pod.Name)
			return err
		}
		log.Info("Deleted pod to restart", "name", pod.Name)
	}
	return nil
}
