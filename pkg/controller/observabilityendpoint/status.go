// Copyright (c) 2020 Red Hat, Inc.

package observabilityendpoint

import (
	"context"
	"time"

	addonv1alpha1 "github.com/open-cluster-management/addon-framework/api/v1alpha1"
	oav1beta1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/observability/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	conditions = map[string]map[string]string{
		"Ready": map[string]string{
			"type":    "Available",
			"reason":  "Deployed",
			"message": "Metrics collector deployed and functional"},
		"Disabled": map[string]string{
			"type":    "Disabled",
			"reason":  "Disabled",
			"message": "enableMetrics is set to False"},
		"Degraded": map[string]string{
			"type":    "Degraded",
			"reason":  "Degraded",
			"message": "Metrics collector deployment not successful"},
		"NotSupported": map[string]string{
			"type":    "NotSupported",
			"reason":  "NotSupported",
			"message": "Observability is not supported in this cluster"},
	}
)

func reportStatus(c client.Client, i *oav1beta1.ObservabilityAddon, t string) {
	i.Status.Conditions = []oav1beta1.StatusCondition{
		{
			Type:               t,
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             conditions[t]["reason"],
			Message:            conditions[t]["message"],
		},
	}
	err := c.Status().Update(context.TODO(), i)
	if err != nil {
		log.Error(err, "Failed to update status for observabilityaddon")
	}
}

func reportStatusToMCAddon(c client.Client, i *addonv1alpha1.ManagedClusterAddOn, t string) {
	if t == "NotSupported" {
		i.Status.Conditions = []addonv1alpha1.Condition{
			{
				Type:               t,
				Status:             metav1.ConditionTrue,
				LastTransitionTime: metav1.NewTime(time.Now()),
				Reason:             conditions[t]["reason"],
				Message:            conditions[t]["message"],
			},
		}
	} else {
		// If Supported Change status for type Available, Degraded and Disabled
		var conditionArray = make([]addonv1alpha1.Condition, 3)
		count := 0
		for key := range conditions {
			if key == "NotSupported" {
				continue
			}
			var status metav1.ConditionStatus
			if key == t {
				status = metav1.ConditionTrue
			} else {
				status = metav1.ConditionFalse
			}
			conditionArray[count] = addonv1alpha1.Condition{
				Type:               conditions[key]["type"],
				Status:             status,
				LastTransitionTime: metav1.NewTime(time.Now()),
				Reason:             conditions[key]["reason"],
				Message:            conditions[key]["message"],
			}
			count++
		}
		i.Status.Conditions = conditionArray
	}

	err := c.Status().Update(context.TODO(), i)
	if err != nil {
		log.Error(err, "Failed to update status for mamagedclusteraddon")
	}
}
