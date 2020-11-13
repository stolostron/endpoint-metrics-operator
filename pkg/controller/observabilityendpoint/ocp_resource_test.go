// Copyright (c) 2020 Red Hat, Inc.

package observabilityendpoint

import (
	"testing"

	ocinfrav1 "github.com/openshift/api/config/v1"
	fakeocpclient "github.com/openshift/client-go/config/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

func TestCreateDeleteCAConfigmap(t *testing.T) {
	c := fake.NewFakeClient()
	err := createCAConfigmap(c)
	if err != nil {
		t.Fatalf("Failed to create CA configmap: (%v)", err)
	}
	err = deleteCAConfigmap(c)
	if err != nil {
		t.Fatalf("Failed to delete CA configmap: (%v)", err)
	}
}

func TestCreateDeleteMonitoringClusterRoleBinding(t *testing.T) {
	c := fake.NewFakeClient()
	err := createMonitoringClusterRoleBinding(c)
	if err != nil {
		t.Fatalf("Failed to create clusterrolebinding: (%v)", err)
	}
	err = deleteMonitoringClusterRoleBinding(c)

	if err != nil {
		t.Fatalf("Failed to delete clusterrolebinding: (%v)", err)
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
