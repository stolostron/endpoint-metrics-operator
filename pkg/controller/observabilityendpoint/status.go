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
		"NotSupported": map[string]string{
			"type":    "Available",
			"reason":  "NotSupported",
			"message": "Observability is not supported in this cluster"},
		"Degraded": map[string]string{
			"type":    "Degraded",
			"reason":  "Degraded",
			"message": "Metrics collector deployment not successful"},
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

	i.Status.Conditions = []addonv1alpha1.Condition{
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
		log.Error(err, "Failed to update status for mamagedclusteraddon")
	}
}

/*
// createCondition returns a condition based on given information
func createCondition(
	conditionType string,
	status metav1.ConditionStatus,
	reason string,
	msg string,
) *addonv1alpha1.Condition {
	return &addonv1alpha1.Condition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             reason,
		Message:            msg,
	}
}

// setStatusCondition appends new if there is no existed condition with same type
// will override a condition if it is with the same type, will do no changes if type & status & reason are the same
// this method assumes the given array of conditions don't have any two conditions with the same type
func setStatusCondition(conditions *[]addonv1alpha1.Condition, condition *addonv1alpha1.Condition) {
	for i, c := range *conditions {
		if c.Type == condition.Type {
			if c.Status != condition.Status || c.Reason != condition.Reason {
				(*conditions)[i] = *condition
			}
			return
		}
	}
	*conditions = append(*conditions, *condition)
}
*/
