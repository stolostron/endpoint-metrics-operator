// Copyright (c) 2020 Red Hat, Inc.

package observabilityendpoint

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
	name          = "observability-addon"
	testNamespace = "test-ns"
)

func newObservabilityAddon() *oav1beta1.ObservabilityAddon {
	return &oav1beta1.ObservabilityAddon{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
		},
		Spec: oav1beta1.ObservabilityAddonSpec{
			EnableMetrics: true,
			MetricsConfigs: oav1beta1.MetricsConfigsSpec{
				Interval: "1m",
			},
		},
	}
}

func TestObservabilityAddonController(t *testing.T) {

	s := scheme.Scheme
	if err := apis.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add oav1beta1 scheme: (%v)", err)
	}

	oa := newObservabilityAddon()
	objs := []runtime.Object{oa, hubInfo}

	kubeClient := kubefakeclient.NewSimpleClientset([]runtime.Object{}...)
	ocpClient := fakeconfigclient.NewSimpleClientset(cv)

	s.AddKnownTypes(oav1beta1.SchemeGroupVersion, oa)
	c := fake.NewFakeClient(objs...)
	r := &ReconcileObservabilityAddon{
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
	r = &ReconcileObservabilityAddon{
		client:     c,
		scheme:     s,
		kubeClient: kubeClient,
		ocpClient:  ocpClient,
	}
	req = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: testNamespace,
		},
	}
	_, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	oa.Spec.EnableMetrics = false
	c = fake.NewFakeClient(objs...)
	r = &ReconcileObservabilityAddon{
		client:     c,
		scheme:     s,
		kubeClient: kubeClient,
		ocpClient:  ocpClient,
	}
	req = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: testNamespace,
		},
	}
	_, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("2nd reconcile: (%v)", err)
	}
}

func TestObservabilityAddonControllerFinalizer(t *testing.T) {

	s := scheme.Scheme

	if err := apis.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add oav1beta1 scheme: (%v)", err)
	}

	oa := newObservabilityAddon()
	oa.ObjectMeta.DeletionTimestamp = &v1.Time{time.Now()}
	oa.ObjectMeta.Finalizers = []string{epFinalizer, "test-finalizerr"}

	objs := []runtime.Object{oa, hubInfo}
	s.AddKnownTypes(oav1beta1.SchemeGroupVersion, oa)
	c := fake.NewFakeClient(objs...)
	kubeClient := kubefakeclient.NewSimpleClientset([]runtime.Object{}...)
	ocpClient := fakeconfigclient.NewSimpleClientset(cv)
	r := &ReconcileObservabilityAddon{
		client:     c,
		scheme:     s,
		kubeClient: kubeClient,
		ocpClient:  ocpClient,
	}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: testNamespace,
		},
	}
	_, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
}
