// Copyright (c) 2021 Red Hat, Inc.

package status

import (
	"context"
	"os"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/open-cluster-management/endpoint-metrics-operator/pkg/util"
	oav1beta1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/observability/v1beta1"
)

const (
	hubKubeConfigPath       = "/spoke/hub-kubeconfig/kubeconfig"
	hubConfigName           = "hub-info-secret"
	obAddonName             = "observability-addon"
	mcoCRName               = "observability"
	ownerLabelKey           = "owner"
	ownerLabelValue         = "multicluster-operator"
	obsAddonFinalizer       = "observability.open-cluster-management.io/addon-cleanup"
	managedClusterAddonName = "observability-controller"
)

var (
	namespace    = os.Getenv("WATCH_NAMESPACE")
	hubNamespace = os.Getenv("HUB_NAMESPACE")
	log          = logf.Log.WithName("controller_status")
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Status Controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager
// is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	// Create kube client
	kubeClient, err := util.CreateHubClient()
	if err != nil {
		log.Error(err, "Failed to create the Kubernetes client")
		return nil
	}
	return &ReconcileStatus{
		client:    mgr.GetClient(),
		hubClient: kubeClient,
		scheme:    mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	if os.Getenv("NAMESPACE") != "" {
		namespace = os.Getenv("NAMESPACE")
	}
	// Create a new controller
	c, err := controller.New("status-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.MetaNew.GetNamespace() == namespace &&
				!reflect.DeepEqual(e.ObjectNew.(*oav1beta1.ObservabilityAddon).Status,
					e.ObjectOld.(*oav1beta1.ObservabilityAddon).Status) {
				return true
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}

	err = c.Watch(&source.Kind{Type: &oav1beta1.ObservabilityAddon{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileObservabilityAddon implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileStatus{}

// ReconcileStatus reconciles a ObservabilityAddon object
type ReconcileStatus struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	scheme    *runtime.Scheme
	hubClient client.Client
}

// Reconcile reads that state of the cluster for a ObservabilityAddon object and makes changes based on the state read
// and what is in the ObservabilityAddon.Status
// Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileStatus) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling")

	// Fetch the ObservabilityAddon instance in hub cluster
	hubObsAddon := &oav1beta1.ObservabilityAddon{}
	err := r.hubClient.Get(context.TODO(), types.NamespacedName{Name: obAddonName, Namespace: hubNamespace}, hubObsAddon)
	if err != nil {
		log.Error(err, "Failed to get observabilityaddon in hub cluster", "namespace", hubNamespace)
		return reconcile.Result{}, err
	}

	// Fetch the ObservabilityAddon instance in local cluster
	obsAddon := &oav1beta1.ObservabilityAddon{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: obAddonName, Namespace: namespace}, obsAddon)
	if err != nil {
		log.Error(err, "Failed to get observabilityaddon", "namespace", namespace)
		return reconcile.Result{}, err
	}

	hubObsAddon.Status = obsAddon.Status

	err = r.hubClient.Status().Update(context.TODO(), hubObsAddon)
	if err != nil {
		log.Error(err, "Failed to update status for observabilityaddon in hub cluster", "namespace", hubNamespace)
	}

	return reconcile.Result{}, nil
}
