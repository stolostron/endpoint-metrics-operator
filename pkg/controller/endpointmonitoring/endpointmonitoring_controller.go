// Copyright (c) 2020 Red Hat, Inc.

package endpointmonitoring

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	monitoringv1alpha1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/open-cluster-management/multicluster-monitoring-operator/pkg/controller/util"
)

const (
	epConfigName    = "endpoint-config"
	ownerLabelKey   = "owner"
	ownerLabelValue = "multicluster-operator"
	epFinalizer     = "monitoring.open-cluster-management.io/endpoint-monitoring-cleanup"
)

var log = logf.Log.WithName("controller_endpointmonitoring")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new EndpointMonitoring Controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager
// is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileEndpointMonitoring{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("endpointmonitoring-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Meta.GetName() == epConfigName && e.Meta.GetAnnotations()[ownerLabelKey] == ownerLabelValue {
				return true
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.MetaNew.GetName() == epConfigName && e.MetaNew.GetAnnotations()[ownerLabelKey] == ownerLabelValue {
				return true
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if e.Meta.GetName() == epConfigName && e.Meta.GetAnnotations()[ownerLabelKey] == ownerLabelValue {
				return !e.DeleteStateUnknown
			}
			return false
		},
	}

	// Watch for changes to primary resource EndpointMonitoring
	err = c.Watch(&source.Kind{Type: &monitoringv1alpha1.EndpointMonitoring{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileEndpointMonitoring implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileEndpointMonitoring{}

// ReconcileEndpointMonitoring reconciles a EndpointMonitoring object
type ReconcileEndpointMonitoring struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a EndpointMonitoring object and makes changes based on the state read
// and what is in the EndpointMonitoring.Spec
// Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileEndpointMonitoring) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling EndpointMonitoring")

	// Fetch the EndpointMonitoring instance
	instance := &monitoringv1alpha1.EndpointMonitoring{}
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

	for _, collector := range instance.Spec.MetricsCollectorList {
		if collector.Type == "OCP_PROMETHEUS" {
			err = util.UpdateClusterMonitoringConfig(instance.Spec.GlobalConfig.SeverURL, &collector.RelabelConfigs)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else {
			reqLogger.Info("Unsupported collector", "type", collector.Type)
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileEndpointMonitoring) initFinalization(ep *monitoringv1alpha1.EndpointMonitoring) error {
	if ep.GetDeletionTimestamp() != nil && contains(ep.GetFinalizers(), epFinalizer) {
		log.Info("To revert configurations")
		for _, collector := range ep.Spec.MetricsCollectorList {
			if collector.Type == "OCP_PROMETHEUS" {
				err := util.UpdateClusterMonitoringConfig(ep.Spec.GlobalConfig.SeverURL, nil)
				if err != nil {
					return err
				}
			} else {
				log.Info("Unsupported collector", "type", collector.Type)
			}
		}
		ep.SetFinalizers(remove(ep.GetFinalizers(), epFinalizer))
		err := r.client.Update(context.TODO(), ep)
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
