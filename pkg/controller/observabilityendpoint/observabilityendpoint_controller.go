// Copyright (c) 2020 Red Hat, Inc.

package observabilityendpoint

import (
	"context"
	"os"

	ocpClientSet "github.com/openshift/client-go/config/clientset/versioned"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	"github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis"
	oav1beta1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/observability/v1beta1"
)

const (
	hubKubeConfigPath       = "/spoke/hub-kubeconfig/kubeconfig"
	hubConfigName           = "hub-info-secret"
	obAddonName             = "observability-addon"
	mcoCRName               = "observability"
	ownerLabelKey           = "owner"
	ownerLabelValue         = "multicluster-operator"
	epFinalizer             = "observability.open-cluster-management.io/addon-cleanup"
	managedClusterAddonName = "observability-controller"
	promSvcName             = "prometheus-k8s"
	promNamespace           = "openshift-monitoring"
)

var (
	namespace    = os.Getenv("WATCH_NAMESPACE")
	hubNamespace = os.Getenv("HUB_NAMESPACE")
	log          = logf.Log.WithName("controller_observabilityaddon")
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ObservabilityAddon Controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager
// is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	// Create kube client
	kubeClient, err := createHubClient()
	if err != nil {
		log.Error(err, "Failed to create the Kubernetes client")
		return nil
	}
	// Create OCP client
	ocpClient, err := createOCPClient()
	if err != nil {
		log.Error(err, "Failed to create the OpenShift client")
		return nil
	}
	return &ReconcileObservabilityAddon{
		client:    mgr.GetClient(),
		hubClient: kubeClient,
		ocpClient: ocpClient,
		scheme:    mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("endpointmonitoring-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = watchHubInfo(c)
	if err != nil {
		return err
	}

	err = watchCertificates(c)
	if err != nil {
		return err
	}

	err = watchWhitelist(c)
	if err != nil {
		return err
	}

	err = watchMetricsCollector(c)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileObservabilityAddon implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileObservabilityAddon{}

// ReconcileObservabilityAddon reconciles a ObservabilityAddon object
type ReconcileObservabilityAddon struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	scheme    *runtime.Scheme
	hubClient client.Client
	ocpClient ocpClientSet.Interface
}

// Reconcile reads that state of the cluster for a ObservabilityAddon object and makes changes based on the state read
// and what is in the ObservabilityAddon.Spec
// Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileObservabilityAddon) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling")

	// Fetch the ObservabilityAddon instance
	obsAddon := &oav1beta1.ObservabilityAddon{}
	err := r.hubClient.Get(context.TODO(), types.NamespacedName{Name: obAddonName, Namespace: hubNamespace}, obsAddon)
	if err != nil {
		log.Error(err, "Failed to get observabilityaddon", "namespace", hubNamespace)
		return reconcile.Result{}, err
	}

	hubSecret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: hubConfigName, Namespace: namespace}, hubSecret)
	if err != nil {
		return reconcile.Result{}, err
	}
	hubInfo := &HubInfo{}
	err = yaml.Unmarshal(hubSecret.Data[hubInfoKey], &hubInfo)
	if err != nil {
		log.Error(err, "Failed to unmarshal hub info")
		return reconcile.Result{}, err
	}

	// Init finalizers
	deleted, err := r.initFinalization(*hubInfo, obsAddon)
	if err != nil {
		return reconcile.Result{}, err
	}
	if deleted {
		return reconcile.Result{}, nil
	}

	// Fetch the ManagedClusterAddon instance
	mcaInstance := &addonv1alpha1.ManagedClusterAddOn{}
	err = r.hubClient.Get(context.TODO(), types.NamespacedName{Name: managedClusterAddonName,
		Namespace: hubNamespace}, mcaInstance)
	if err != nil {
		log.Error(err, "Failed to get managedclusteraddon", "namespace", hubNamespace)
		return reconcile.Result{}, err

	}

	// If no prometheus service found, set status as NotSupported
	promSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: promSvcName,
		Namespace: promNamespace}, promSvc)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Error(err, "OCP prometheus service does not exist")
			reportStatus(r.hubClient, obsAddon, "NotSupported")
			reportStatusToMCAddon(r.hubClient, mcaInstance, "NotSupported")
			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "Failed to check prometheus resource")
		return reconcile.Result{}, err
	}

	clusterID, err := getClusterID(r.ocpClient)
	if err != nil {
		// OCP 3.11 has no cluster id, set it as empty string
		clusterID = ""
	}

	err = createMonitoringClusterRoleBinding(r.client)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = createCAConfigmap(r.client)
	if err != nil {
		return reconcile.Result{}, err
	}

	if hubInfo.EnableMetrics {
		forceRestart := false
		if request.Name == mtlsCertName {
			forceRestart = true
		}
		created, err := updateMetricsCollector(r.client, *hubInfo, clusterID, 1, forceRestart)
		if err != nil {
			reportStatusToMCAddon(r.client, mcaInstance, "Degraded")
			return reconcile.Result{}, err
		}
		if created {
			reportStatus(r.hubClient, obsAddon, "Ready")
			reportStatusToMCAddon(r.hubClient, mcaInstance, "Ready")
		}
	} else {
		deleted, err := updateMetricsCollector(r.client, *hubInfo, clusterID, 0, false)
		if err != nil {
			return reconcile.Result{}, err
		}
		if deleted {
			reportStatus(r.hubClient, obsAddon, "Disabled")
			reportStatusToMCAddon(r.hubClient, mcaInstance, "Disabled")
		}
	}

	//TODO: UPDATE
	return reconcile.Result{}, nil
}

func (r *ReconcileObservabilityAddon) initFinalization(
	hubInfo HubInfo, ep *oav1beta1.ObservabilityAddon) (bool, error) {
	if hubInfo.DeleteFlag && contains(ep.GetFinalizers(), epFinalizer) {
		log.Info("To clean observability components/configurations in the cluster")
		err := deleteMetricsCollector(r.client)
		if err != nil {
			return false, err
		}
		// Should we return bool from the delete functions for crb and cm? What is it used for? Should we use the bool before removing finalizer?
		//SHould we return true if metricscollector is not found as that means  metrics collector is not present?
		//Moved this part up as we need to clean up cm and crb before we remove the finalizer - is that the right way to do it?
		err = deleteMonitoringClusterRoleBinding(r.client)
		if err != nil {
			return false, err
		}
		err = deleteCAConfigmap(r.client)
		if err != nil {
			return false, err
		}
		ep.SetFinalizers(remove(ep.GetFinalizers(), epFinalizer))
		err = r.hubClient.Update(context.TODO(), ep)
		if err != nil {
			log.Error(err, "Failed to remove finalizer to observabilityaddon", "namespace", ep.Namespace)
			return false, err
		}
		log.Info("Finalizer removed from observabilityaddon resource")
		return true, nil
	}
	if !contains(ep.GetFinalizers(), epFinalizer) {
		ep.SetFinalizers(append(ep.GetFinalizers(), epFinalizer))
		err := r.hubClient.Update(context.TODO(), ep)
		if err != nil {
			log.Error(err, "Failed to add finalizer to observabilityaddon", "namespace", ep.Namespace)
			return false, err
		}
		log.Info("Finalizer added to observabilityaddon resource")
	}
	return false, nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func remove(list []string, s string) []string {
	result := []string{}
	for _, v := range list {
		if v != s {
			result = append(result, v)
		}
	}
	return result
}

func createOCPClient() (ocpClientSet.Interface, error) {
	// create the config from the path
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Error(err, "Failed to create the config")
		return nil, err
	}

	// generate the client based off of the config
	ocpClient, err := ocpClientSet.NewForConfig(config)
	if err != nil {
		log.Error(err, "Failed to create ocp config client")
		return nil, err
	}

	return ocpClient, err
}

func createHubClient() (client.Client, error) {
	// create the config from the path
	config, err := clientcmd.BuildConfigFromFlags("", hubKubeConfigPath)
	if err != nil {
		log.Error(err, "Failed to create the config")
		return nil, err
	}

	s := scheme.Scheme
	if err := apis.AddToScheme(s); err != nil {
		return nil, err
	}

	if err := addonv1alpha1.AddToScheme(s); err != nil {
		return nil, err
	}

	// generate the client based off of the config
	hubClient, err := client.New(config, client.Options{Scheme: s})
	if err != nil {
		log.Error(err, "Failed to create hub client")
		return nil, err
	}

	return hubClient, err
}

func watchHubInfo(c controller.Controller) error {
	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Meta.GetName() == hubConfigName && e.Meta.GetNamespace() == namespace {
				return true
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.MetaNew.GetName() == hubConfigName && e.MetaNew.GetNamespace() == namespace &&
				e.MetaNew.GetResourceVersion() != e.MetaOld.GetResourceVersion() {
				return true
			}
			return false
		},
	}
	return c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForObject{}, pred)
}

func watchCertificates(c controller.Controller) error {
	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Meta.GetName() == mtlsCertName && e.Meta.GetNamespace() == namespace {
				return true
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.MetaNew.GetName() == mtlsCertName && e.MetaNew.GetNamespace() == namespace &&
				e.MetaNew.GetResourceVersion() != e.MetaOld.GetResourceVersion() {
				return true
			}
			return false
		},
	}
	return c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForObject{}, pred)

}

func watchWhitelist(c controller.Controller) error {
	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Meta.GetName() == metricsConfigMapName && e.Meta.GetNamespace() == namespace {
				return true
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.MetaNew.GetName() == metricsConfigMapName && e.MetaNew.GetNamespace() == namespace &&
				e.MetaNew.GetResourceVersion() != e.MetaOld.GetResourceVersion() {
				return true
			}
			return false
		},
	}
	return c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForObject{}, pred)
}

func watchMetricsCollector(c controller.Controller) error {
	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Meta.GetName() == metricsCollectorName && e.Meta.GetNamespace() == namespace {
				return true
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.MetaNew.GetName() == metricsCollectorName && e.MetaNew.GetNamespace() == namespace &&
				e.MetaNew.GetResourceVersion() != e.MetaOld.GetResourceVersion() {
				return true
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if e.Meta.GetName() == metricsCollectorName && e.Meta.GetNamespace() == namespace {
				return true
			}
			return false
		},
	}
	return c.Watch(&source.Kind{Type: &v1.Deployment{}}, &handler.EnqueueRequestForObject{}, pred)
}
