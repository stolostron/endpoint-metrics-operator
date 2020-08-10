// Copyright (c) 2020 Red Hat, Inc.

package endpointmonitoring

import (
	"testing"
	"time"

	fakeconfigclient "github.com/openshift/client-go/config/clientset/versioned/fake"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kubefakeclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis"
	oav1beta1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/observability/v1beta1"
)

const (
	name      = "observability-addon"
	namespace = "test-ns"
)

func newEndpoint() *oav1beta1.EndpointMonitoring {
	return &oav1beta1.EndpointMonitoring{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: oav1beta1.EndpointMonitoringSpec{
			GlobalConfig: oav1beta1.GlobalConfigSpec{},
			MetricsCollectorList: []oav1beta1.MetricsCollectorSpec{
				{
					Enable: true,
				},
			},
		},
	}
}

func TestEndpointMonitoringController(t *testing.T) {

	s := scheme.Scheme
	if err := apis.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add oav1beta1 scheme: (%v)", err)
	}

	ep := newEndpoint()
	objs := []runtime.Object{ep}

	kubeClient := kubefakeclient.NewSimpleClientset([]runtime.Object{}...)
	ocpClient := fakeconfigclient.NewSimpleClientset([]runtime.Object{}...)

	s.AddKnownTypes(oav1beta1.SchemeGroupVersion, ep)
	c := fake.NewFakeClient(objs...)
	r := &ReconcileEndpointMonitoring{
		client:     c,
		scheme:     s,
		kubeClient: kubeClient,
		ocpClient:  ocpClient,
	}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	_, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	c = fake.NewFakeClient()
	r = &ReconcileEndpointMonitoring{
		client:     c,
		scheme:     s,
		kubeClient: kubeClient,
		ocpClient:  ocpClient,
	}
	req = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	_, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	ep.Spec.MetricsCollectorList[0].Type = "OCP_PROMETHEUS"
	c = fake.NewFakeClient(objs...)
	r = &ReconcileEndpointMonitoring{
		client:     c,
		scheme:     s,
		kubeClient: kubeClient,
		ocpClient:  ocpClient,
	}
	req = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	_, err = r.Reconcile(req)
	if err == nil {
		t.Fatalf("Desied errors not found")
	}
}

func TestEndpointMonitoringControllerFinalizer(t *testing.T) {

	s := scheme.Scheme

	if err := apis.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add oav1beta1 scheme: (%v)", err)
	}

	ep := newEndpoint()
	ep.ObjectMeta.DeletionTimestamp = &v1.Time{time.Now()}
	ep.ObjectMeta.Finalizers = []string{epFinalizer, "test-finalizerr"}

	objs := []runtime.Object{ep}
	s.AddKnownTypes(oav1beta1.SchemeGroupVersion, ep)
	c := fake.NewFakeClient(objs...)
	kubeClient := kubefakeclient.NewSimpleClientset([]runtime.Object{}...)
	ocpClient := fakeconfigclient.NewSimpleClientset([]runtime.Object{}...)
	r := &ReconcileEndpointMonitoring{
		client:     c,
		scheme:     s,
		kubeClient: kubeClient,
		ocpClient:  ocpClient,
	}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	_, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	ep.Spec.MetricsCollectorList[0].Type = "OCP_PROMETHEUS"
	c = fake.NewFakeClient(objs...)
	r = &ReconcileEndpointMonitoring{
		client:     c,
		scheme:     s,
		kubeClient: kubeClient,
		ocpClient:  ocpClient,
	}
	req = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	_, err = r.Reconcile(req)
	if err == nil {
		t.Fatalf("Desied errors not found")
	}
}
