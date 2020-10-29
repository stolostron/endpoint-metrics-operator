// Copyright (c) 2020 Red Hat, Inc.

package observabilityendpoint

import (
	"testing"
	"time"

	fakeconfigclient "github.com/openshift/client-go/config/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kubefakeclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	"github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis"
	oav1beta1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/observability/v1beta1"
)

const (
	name          = "observability-addon"
	testNamespace = "test-ns"
	hubInfoName   = "hub-info-secret"
)

func newObservabilityAddon() *oav1beta1.ObservabilityAddon {
	return &oav1beta1.ObservabilityAddon{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
		},
	}
}

func newMCOResource() *oav1beta1.MultiClusterObservability {
	return &oav1beta1.MultiClusterObservability{
		ObjectMeta: v1.ObjectMeta{
			Name: mcoCRName,
		},
		Spec: oav1beta1.MultiClusterObservabilitySpec{
			ObservabilityAddonSpec: &oav1beta1.ObservabilityAddonSpec{
				EnableMetrics: true,
				Interval:      60,
			},
		},
	}
}

func newManagedClusterAddon() *addonv1alpha1.ManagedClusterAddOn {
	return &addonv1alpha1.ManagedClusterAddOn{
		TypeMeta: metav1.TypeMeta{
			APIVersion: addonv1alpha1.SchemeGroupVersion.String(),
			Kind:       "ManagedClusterAddOn",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      managedClusterAddonName,
			Namespace: namespace,
		},
	}
}

func newHubInfoSecret() *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hubInfoName,
			Namespace: namespace,
		},
		Data: map[string][]byte{},
	}
}

func TestObservabilityAddonController(t *testing.T) {

	s := scheme.Scheme
	if err := apis.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add oav1beta1 scheme: (%v)", err)
	}

	oa := newObservabilityAddon()
	mcoa := newMCOResource()
	mcaddon := newManagedClusterAddon()
	objs := []runtime.Object{oa, hubInfo, mcoa, mcaddon}

	hubInfo := newHubInfoSecret()

	kubeClient := kubefakeclient.NewSimpleClientset(hubInfo)
	ocpClient := fakeconfigclient.NewSimpleClientset(cv)

	s.AddKnownTypes(oav1beta1.SchemeGroupVersion, oa)
	s.AddKnownTypes(oav1beta1.SchemeGroupVersion, mcoa)
	s.AddKnownTypes(addonv1alpha1.SchemeGroupVersion, mcaddon)
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

	mcoa.Spec.ObservabilityAddonSpec.EnableMetrics = false
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
