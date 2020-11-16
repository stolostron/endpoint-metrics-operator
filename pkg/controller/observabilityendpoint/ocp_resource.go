// Copyright (c) 2020 Red Hat, Inc.

package observabilityendpoint

import (
	"context"
	"os"
	"reflect"

	ocpClientSet "github.com/openshift/client-go/config/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	clusterRoleBindingName = "metrics-collector-view"
	caConfigmapName        = "metrics-collector-serving-certs-ca-bundle"
)

var (
	serviceAccountName = os.Getenv("SERVICE_ACCOUNT")
)

func deleteMonitoringClusterRoleBinding(client client.Client) error {
	rb := &rbacv1.ClusterRoleBinding{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: clusterRoleBindingName,
		Namespace: ""}, rb)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("clusterrolebinding already deleted")
			return nil
		}
		log.Error(err, "Failed to check the clusterrolebinding")
		return err
	}
	err = client.Delete(context.TODO(), rb)
	if err != nil {
		log.Error(err, "Error deleting clusterrolebinding")
		return err
	}
	log.Info("clusterrolebinding deleted")
	return nil
}

func createMonitoringClusterRoleBinding(client client.Client) error {
	rb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
			Annotations: map[string]string{
				ownerLabelKey: ownerLabelValue,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "cluster-monitoring-view",
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: namespace,
			},
		},
	}

	found := &rbacv1.ClusterRoleBinding{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: clusterRoleBindingName,
		Namespace: ""}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			err = client.Create(context.TODO(), rb)
			if err == nil {
				log.Info("clusterrolebinding created")
			} else {
				log.Error(err, "Failed to create the clusterrolebinding")
			}
			return err
		}
		log.Error(err, "Failed to check the clusterrolebinding")
		return err
	}

	if reflect.DeepEqual(rb.RoleRef, found.RoleRef) && reflect.DeepEqual(rb.Subjects, found.Subjects) {
		log.Info("The clusterrolebinding already existed")
	} else {
		rb.ObjectMeta.ResourceVersion = found.ObjectMeta.ResourceVersion
		err = client.Update(context.TODO(), rb)
		if err != nil {
			log.Error(err, "Failed to update the clusterrolebinding")
		}
	}

	return nil
}

func deleteCAConfigmap(client client.Client) error {
	cm := &corev1.ConfigMap{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: caConfigmapName,
		Namespace: namespace}, cm)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("configmap already deleted")
			return nil
		}
		log.Error(err, "Failed to check the configmap")
		return err
	}
	err = client.Delete(context.TODO(), cm)
	if err != nil {
		log.Error(err, "Error deleting configmap")
		return err
	}
	log.Info("configmap deleted")
	return nil
}

func createCAConfigmap(client client.Client) error {
	cm := &corev1.ConfigMap{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: caConfigmapName,
		Namespace: namespace}, cm)
	if err != nil {
		if errors.IsNotFound(err) {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      caConfigmapName,
					Namespace: namespace,
					Annotations: map[string]string{
						ownerLabelKey: ownerLabelValue,
						"service.alpha.openshift.io/inject-cabundle": "true",
					},
				},
				Data: map[string]string{"service-ca.crt": ""},
			}
			err = client.Create(context.TODO(), cm)
			if err == nil {
				log.Info("Configmap created")
			} else {
				log.Error(err, "Failed to create the configmap")
			}
			return err
		} else {
			log.Error(err, "Failed to check the configmap")
			return err
		}
	} else {
		log.Info("The configmap already existed")
	}
	return nil
}

// getClusterID is used to get the cluster uid
func getClusterID(ocpClient ocpClientSet.Interface) (string, error) {
	clusterVersion, err := ocpClient.ConfigV1().ClusterVersions().Get(context.TODO(), "version", metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Failed to get clusterVersion")
		return "", err
	}

	return string(clusterVersion.Spec.ClusterID), nil
}
