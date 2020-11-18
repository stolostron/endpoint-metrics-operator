// Copyright (c) 2020 Red Hat, Inc.

package observabilityendpoint

import (
	"context"
	"strings"
	"testing"

	fakeconfigclient "github.com/openshift/client-go/config/clientset/versioned/fake"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
	podName         = "metrics-collector-deployment-abc-xyz"
)

func newPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels: map[string]string{
				selectorKey: selectorValue,
			},
		},
	}
}

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

func newHubInfoSecret(deleteFlag bool, enableMetrics bool) *corev1.Secret {
	data := []byte(`
cluster-name: "test-cluster"
endpoint: "http://test-endpoint"
enable-metrics: true
internal: 60
delete-flag: false
`)
	if deleteFlag {
		data = []byte(`
cluster-name: "test-cluster"
endpoint: "http://test-endpoint"
enable-metrics: true
internal: 60
delete-flag: true
`)
	} else if !enableMetrics {
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

	oba := newObservabilityAddon()
	mcaddon := newManagedClusterAddon()
	hubObjs := []runtime.Object{}
	hubInfo := newHubInfoSecret(false, true)
	whiteList := getWhitelistCM()
	objs := []runtime.Object{hubInfo, whiteList}

	hubClient := fake.NewFakeClient(hubObjs...)
	ocpClient := fakeconfigclient.NewSimpleClientset(cv)
	c := fake.NewFakeClient(objs...)

	r := &ReconcileObservabilityAddon{
		client:    c,
		hubClient: hubClient,
		ocpClient: ocpClient,
	}

	// test error in reconcile if missing obervabilityaddon
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "install",
			Namespace: testNamespace,
		},
	}
	_, err := r.Reconcile(req)
	if err == nil {
		t.Fatalf("reconcile: miss the error for missing obervabilityaddon")
	}

	// test error in reconcile if missing managedclusteraddon
	err = hubClient.Create(context.TODO(), oba)
	if err != nil {
		t.Fatalf("failed to create oba to install: (%v)", err)
	}
	req = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "install",
			Namespace: testNamespace,
		},
	}
	_, err = r.Reconcile(req)
	if err == nil {
		t.Fatalf("reconcile: miss the error for missing managedclusteraddon")
	}

	// test reconcile w/o prometheus-k8s svc
	err = hubClient.Create(context.TODO(), mcaddon)
	if err != nil {
		t.Fatalf("failed to create mcaddon to install: (%v)", err)
	}
	req = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "install",
			Namespace: testNamespace,
		},
	}
	_, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// test reconcile successfully with all resources installed and finalizer set
	promSvc := newPromSvc()
	err = c.Create(context.TODO(), promSvc)
	if err != nil {
		t.Fatalf("failed to create prom svc to install: (%v)", err)
	}
	req = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "install",
			Namespace: testNamespace,
		},
	}
	_, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	rb := &rbacv1.ClusterRoleBinding{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: clusterRoleBindingName,
		Namespace: ""}, rb)
	if err != nil {
		t.Fatalf("Required clusterrolebinding not created: (%v)", err)
	}
	cm := &corev1.ConfigMap{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: caConfigmapName,
		Namespace: namespace}, cm)
	if err != nil {
		t.Fatalf("Required configmap not created: (%v)", err)
	}
	deploy := &appv1.Deployment{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: metricsCollectorName,
		Namespace: namespace}, deploy)
	if err != nil {
		t.Fatalf("Metrics collector deployment not created: (%v)", err)
	}
	foundOba := &oav1beta1.ObservabilityAddon{}
	err = hubClient.Get(context.TODO(), types.NamespacedName{Name: obAddonName,
		Namespace: hubNamespace}, foundOba)
	if err != nil {
		t.Fatalf("Failed to get observabilityAddon: (%v)", err)
	}
	if !contains(foundOba.Finalizers, epFinalizer) {
		t.Fatal("Finalizer not set in observabilityAddon")
	}

	// test reconcile w/o clusterversion(OCP 3.11)
	r.ocpClient = fakeconfigclient.NewSimpleClientset()
	req = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "install",
			Namespace: testNamespace,
		},
	}
	_, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	err = c.Get(context.TODO(), types.NamespacedName{Name: metricsCollectorName,
		Namespace: namespace}, deploy)
	if err != nil {
		t.Fatalf("Metrics collector deployment not created: (%v)", err)
	}
	commands := deploy.Spec.Template.Spec.Containers[0].Command
	for _, cmd := range commands {
		if strings.Contains(cmd, "clusterID=") && !strings.Contains(cmd, "test-cluster") {
			t.Fatalf("Found wrong clusterID in command: (%s)", cmd)
		}
	}

	// test reconcile metrics collector pod deleted if cert secret updated
	err = c.Create(context.TODO(), newPod())
	if err != nil {
		t.Fatalf("Failed to create pod: (%v)", err)
	}
	req = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      mtlsCertName,
			Namespace: testNamespace,
		},
	}
	_, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile for update: (%v)", err)
	}
	pod := &corev1.Pod{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: podName,
		Namespace: namespace}, pod)
	if !errors.IsNotFound(err) {
		t.Fatal("Pod not deleted")
	}

	// test reconcile  metrics collector's replicas set to 0 if observability disabled
	err = c.Delete(context.TODO(), hubInfo)
	if err != nil {
		t.Fatalf("failed to delete hubinfo secret to disable: (%v)", err)
	}
	hubInfo = newHubInfoSecret(false, false)
	err = c.Create(context.TODO(), hubInfo)
	if err != nil {
		t.Fatalf("failed to create hubinfo secret to disable: (%v)", err)
	}
	req = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "disable",
			Namespace: testNamespace,
		},
	}
	_, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile for disable: (%v)", err)
	}
	err = c.Get(context.TODO(), types.NamespacedName{Name: metricsCollectorName,
		Namespace: namespace}, deploy)
	if err != nil {
		t.Fatalf("Metrics collector deployment not created: (%v)", err)
	}
	if *deploy.Spec.Replicas != 0 {
		t.Fatalf("Replicas for metrics collector deployment is not set as 0, value is (%d)", *deploy.Spec.Replicas)
	}

	// test reconcile all resources and finalizer are removed
	err = c.Delete(context.TODO(), hubInfo)
	if err != nil {
		t.Fatalf("failed to delete hubinfo secret to disable: (%v)", err)
	}
	hubInfo = newHubInfoSecret(true, true)
	err = c.Create(context.TODO(), hubInfo)
	if err != nil {
		t.Fatalf("failed to create hubinfo secret to disable: (%v)", err)
	}
	if err != nil {
		t.Fatalf("failed to update hubinfo secret to delete: (%v)", err)
	}
	req = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "delete",
			Namespace: testNamespace,
		},
	}
	_, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile for delete: (%v)", err)
	}
	err = c.Get(context.TODO(), types.NamespacedName{Name: clusterRoleBindingName,
		Namespace: ""}, rb)
	if !errors.IsNotFound(err) {
		t.Fatalf("Required clusterrolebinding not deleted")
	}
	err = c.Get(context.TODO(), types.NamespacedName{Name: caConfigmapName,
		Namespace: namespace}, cm)
	if !errors.IsNotFound(err) {
		t.Fatalf("Required configmap not deleted")
	}
	err = c.Get(context.TODO(), types.NamespacedName{Name: metricsCollectorName,
		Namespace: namespace}, deploy)
	if !errors.IsNotFound(err) {
		t.Fatalf("Metrics collector deployment not deleted")
	}
	foundOba1 := &oav1beta1.ObservabilityAddon{}
	err = hubClient.Get(context.TODO(), types.NamespacedName{Name: obAddonName,
		Namespace: hubNamespace}, foundOba1)
	if err != nil {
		t.Fatalf("Failed to get observabilityAddon: (%v)", err)
	}
	if contains(foundOba1.Finalizers, epFinalizer) {
		t.Fatal("Finalizer not removed from observabilityAddon")
	}
}

func TestCreateOCPClient(t *testing.T) {
	_, err := createOCPClient()
	if err == nil {
		t.Fatalf("Failed to catch error when creating ocpclient: (%v)", err)
	}
}

func TestCreateHubClient(t *testing.T) {
	_, err := createHubClient()
	if err == nil {
		t.Fatalf("Failed to catch error when creating hubclient: (%v)", err)
	}
}
