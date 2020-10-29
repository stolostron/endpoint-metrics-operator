// Copyright (c) 2020 Red Hat, Inc.

package observabilityendpoint

import (
	"context"
	"os"

	ocpClientSet "github.com/openshift/client-go/config/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	clusterRoleBindingName = "metrics-collector-view"
	caConfigmapName        = "metrics-collector-serving-certs-ca-bundle"
)

var (
	serviceAccountName = os.Getenv("SERVICE_ACCOUNT")
)

func deleteMonitoringClusterRoleBinding(client kubernetes.Interface) error {
	rb, err := client.RbacV1().ClusterRoleBindings().Get(context.TODO(), clusterRoleBindingName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(err.Error(), ": clusterrolebinding already deleted")
			return nil
		}
		log.Error(err, ": Failed to check the clusterrolebinding")
		return err
	}
	err = client.RbacV1().ClusterRoleBindings().Delete(context.TODO(), rb.GetName(), metav1.DeleteOptions{})
	if err != nil {
		log.Error(err, ": Error deleting clusterrolebinding")
		return err
	}
	log.Info("clusterrolebinding deleted")
	return nil
}

func createMonitoringClusterRoleBinding(client kubernetes.Interface) error {
	_, err := client.RbacV1().ClusterRoleBindings().Get(context.TODO(), clusterRoleBindingName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			rb := &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterRoleBindingName,
					Namespace: namespace,
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
			_, err = client.RbacV1().ClusterRoleBindings().Create(context.TODO(), rb, metav1.CreateOptions{})
			if err == nil {
				log.Info("clusterrolebinding created")
			} else {
				log.Error(err, "Failed to create the clusterrolebinding")
			}
			return err
		}
		log.Error(err, "Failed to check the clusterrolebinding")
		return err
	} else {
		log.Info("The clusterrolebinding already existed")
	}
	return nil
}

func deleteCAConfigmap(client kubernetes.Interface) error {
	//TBD
	cm, err := client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), caConfigmapName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(err.Error(), ": configmap already deleted")
			return nil
		}
		log.Error(err, ": Failed to check the configmap")
		return err
	}
	err = client.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), cm.GetName(), metav1.DeleteOptions{})
	if err != nil {
		log.Error(err, ": Error deleting configmap")
		return err
	}
	log.Info("configmap deleted")
	return nil
}

func createCAConfigmap(client kubernetes.Interface) error {
	_, err := client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), caConfigmapName, metav1.GetOptions{})
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
			_, err = client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
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
