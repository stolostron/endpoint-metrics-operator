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
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	"github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis"
	oav1beta1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/observability/v1beta1"
)

const (
	name            = "observability-addon"
	testNamespace   = "test-ns"
	testHubNamspace = "test-hub-ns"
	hubInfoName     = "hub-info-secret"
)

func newPromSvc() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      promSvcName,
			Namespace: promNamespace,
		},
	}
}

func newObservabilityAddon() *oav1beta1.ObservabilityAddon {
	return &oav1beta1.ObservabilityAddon{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: testHubNamspace,
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
			Namespace: testHubNamspace,
		},
	}
}

func newHubInfoSecret(enableMetrics bool) *corev1.Secret {
	data := []byte(`
cluster-name: "test-cluster"
endpoint: "http://test-endpoint"
enable-metrics: true
internal: 60
delete-flag: false
`)
	if !enableMetrics {
		data = []byte(`
cluster-name: "test-cluster"
endpoint: "http://test-endpoint"
enable-metrics: false
internal: 60
delete-flag: false
`)
	}
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hubConfigName,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			hubInfoKey: data,
		},
	}
}

func init() {
	logf.SetLogger(logf.ZapLogger(true))

	s := scheme.Scheme
	addonv1alpha1.AddToScheme(s)
	apis.AddToScheme(s)

	namespace = testNamespace
	hubNamespace = testHubNamspace
}

func TestObservabilityAddonController(t *testing.T) {

	oa := newObservabilityAddon()
	mcaddon := newManagedClusterAddon()
	hubObjs := []runtime.Object{oa, mcaddon}
	objs := []runtime.Object{newHubInfoSecret(true), newPromSvc(), getWhitelistCM()}

	hubClient := fake.NewFakeClient(hubObjs...)
	ocpClient := fakeconfigclient.NewSimpleClientset(cv)
	c := fake.NewFakeClient(objs...)

	r := &ReconcileObservabilityAddon{
		client:    c,
		hubClient: hubClient,
		ocpClient: ocpClient,
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

	objs = []runtime.Object{newHubInfoSecret(false)}
	c = fake.NewFakeClient(objs...)
	r = &ReconcileObservabilityAddon{
		client:    c,
		hubClient: hubClient,
		ocpClient: ocpClient,
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

	oa := newObservabilityAddon()
	oa.ObjectMeta.DeletionTimestamp = &v1.Time{time.Now()}
	oa.ObjectMeta.Finalizers = []string{epFinalizer, "test-finalizerr"}
	mcaddon := newManagedClusterAddon()

	hubObjs := []runtime.Object{oa, mcaddon}
	objs := []runtime.Object{newHubInfoSecret(true), newPromSvc(), getWhitelistCM()}
	c := fake.NewFakeClient(objs...)
	hubClient := fake.NewFakeClient(hubObjs...)
	ocpClient := fakeconfigclient.NewSimpleClientset(cv)
	r := &ReconcileObservabilityAddon{
		client:    c,
		hubClient: hubClient,
		ocpClient: ocpClient,
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
