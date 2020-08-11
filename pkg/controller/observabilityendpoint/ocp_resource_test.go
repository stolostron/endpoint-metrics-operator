// Copyright (c) 2020 Red Hat, Inc.

package observabilityendpoint

import (
	"testing"

	ocinfrav1 "github.com/openshift/api/config/v1"
	fakeocpclient "github.com/openshift/client-go/config/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	testClusterID = "123456789"
)

var (
	cv = &ocinfrav1.ClusterVersion{
		ObjectMeta: metav1.ObjectMeta{Name: "version"},
		Spec: ocinfrav1.ClusterVersionSpec{
			ClusterID: testClusterID,
		},
	}
)

func TestCreateCAConfigmap(t *testing.T) {
	kubeClient := fake.NewSimpleClientset([]runtime.Object{}...)
	err := createCAConfigmap(kubeClient)
	if err != nil {
		t.Fatalf("Failed to create CA configmap: (%v)", err)
	}
}

func TestCreateMonitoringClusterRoleBinding(t *testing.T) {
	kubeClient := fake.NewSimpleClientset([]runtime.Object{}...)
	err := createMonitoringClusterRoleBinding(kubeClient)
	if err != nil {
		t.Fatalf("Failed to create clusterrolebinding: (%v)", err)
	}
}

func TestGetClusterID(t *testing.T) {
	c := fakeocpclient.NewSimpleClientset(cv)
	found, err := getClusterID(c)
	if err != nil {
		t.Fatalf("Failed to get clusterversion: (%v)", err)
	}
	if found != testClusterID {
		t.Fatalf("Got wrong cluster id" + found)
	}
}
