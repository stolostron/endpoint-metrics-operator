// Copyright (c) 2020 Red Hat, Inc.

package observabilityendpoint

import (
	"context"
	"os"

	ocpClientSet "github.com/openshift/client-go/config/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	oav1beta1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/observability/v1beta1"
)

const (
	hubConfigName   = "hub-info-secret"
	obAddonName     = "observability-addon"
	mcoCRName       = "observability"
	ownerLabelKey   = "owner"
	ownerLabelValue = "multicluster-operator"
	epFinalizer     = "observability.open-cluster-management.io/addon-cleanup"
)

var (
	namespace = os.Getenv("NAMESPACE")
	log       = logf.Log.WithName("controller_observabilityaddon")
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
	kubeClient, err := createKubeClient()
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
		client:     mgr.GetClient(),
		kubeClient: kubeClient,
		ocpClient:  ocpClient,
		scheme:     mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("endpointmonitoring-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		// ManifestWork installs the Controller, so endpoint controller only should care about delete events
		DeleteFunc: func(e event.DeleteEvent) bool {
			if e.Meta.GetName() == obAddonName && e.Meta.GetAnnotations()[ownerLabelKey] == ownerLabelValue {
				return !e.DeleteStateUnknown
			}
			return false
		},
	}

	mcoPred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Meta.GetName() == mcoCRName {
				return true
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.MetaNew.GetName() == mcoCRName {
				return true
			}
			return false
		},
	}

	// Watch for changes to primary resource ObservabilityAddon
	err = c.Watch(&source.Kind{Type: &oav1beta1.ObservabilityAddon{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MCO CR
	err = c.Watch(&source.Kind{Type: &oav1beta1.MultiClusterObservability{}}, &handler.EnqueueRequestForObject{}, mcoPred)
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
	client     client.Client
	scheme     *runtime.Scheme
	kubeClient kubernetes.Interface
	ocpClient  ocpClientSet.Interface
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
	reqLogger.Info("Reconciling ObservabilityAddon")

	// Fetch the ObservabilityAddon instance
	instance := &oav1beta1.ObservabilityAddon{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Init finalizers
	err = r.initFinalization(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	mcoInstance := &oav1beta1.MultiClusterObservability{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: mcoCRName, Namespace: ""}, mcoInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	hubSecret := &corev1.Secret{}
	// hubSecret is in ManifestWork, Read from local k8s client
	// ocp_resource.go
	//	err = r.client.Get(context.TODO(), types.NamespacedName{Name: hubConfigName, Namespace: request.Namespace}, hubSecret)
	hubSecret, err = r.kubeClient.CoreV1().Secrets(namespace).Get(hubConfigName, metav1.GetOptions{}) //(context.TODO(), types.NamespacedName{Name: hubConfigName, Namespace: request.Namespace}, hubSecret)
	if err != nil {
		reqLogger.Error(err, "Failed to get hub secret")
		return reconcile.Result{}, err
	}
	clusterID, err := getClusterID(r.ocpClient)
	if err != nil {
		reportStatus(r.client, instance, "NotSupported")
		return reconcile.Result{}, err
	}

	if mcoInstance.Spec.ObservabilityAddonSpec.EnableMetrics {
		err = createMonitoringClusterRoleBinding(r.kubeClient)
		if err != nil {
			return reconcile.Result{}, err
		}
		err = createCAConfigmap(r.kubeClient)
		if err != nil {
			return reconcile.Result{}, err
		}
		created, err := createMetricsCollector(r.kubeClient, hubSecret, clusterID, &mcoInstance.Spec.ObservabilityAddonSpec)
		if err != nil {
			return reconcile.Result{}, err
		}
		if created {
			reportStatus(r.client, instance, "Ready")
		}
	} else {
		deleted, err := deleteMetricsCollector(r.kubeClient)
		if err != nil {
			return reconcile.Result{}, err
		}
		if deleted {
			reportStatus(r.client, instance, "Disabled")
		}
	}

	//TODO: UPDATE
	return reconcile.Result{}, nil
}

func (r *ReconcileObservabilityAddon) initFinalization(
	ep *oav1beta1.ObservabilityAddon) error {
	if ep.GetDeletionTimestamp() != nil && contains(ep.GetFinalizers(), epFinalizer) {
		log.Info("To revert configurations")
		_, err := deleteMetricsCollector(r.kubeClient)
		if err != nil {
			return err
		}
		ep.SetFinalizers(remove(ep.GetFinalizers(), epFinalizer))
		err = r.client.Update(context.TODO(), ep)
		if err != nil {
			log.Error(err, "Failed to remove finalizer to endpointmonitoring", "namespace", ep.Namespace)
			return err
		}
		log.Info("Finalizer removed from endpointmonitoring resource")
	}
	if !contains(ep.GetFinalizers(), epFinalizer) {
		ep.SetFinalizers(append(ep.GetFinalizers(), epFinalizer))
		err := r.client.Update(context.TODO(), ep)
		if err != nil {
			log.Error(err, "Failed to add finalizer to endpointmonitoring", "namespace", ep.Namespace)
			return err
		}
		log.Info("Finalizer added to endpointmonitoring resource")
	}
	return nil
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

func createKubeClient() (kubernetes.Interface, error) {
	// create the config from the path
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Error(err, "Failed to create the config")
		return nil, err
	}

	// generate the client based off of the config
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err, "Failed to create kube client")
		return nil, err
	}

	return kubeClient, err
}
