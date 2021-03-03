// Copyright (c) 2021 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project.
package util

import (
	"context"
	"time"

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
			"message": "No Prometheus service found in this cluster"},
	}
)

func ReportStatus(c client.Client, i *oav1beta1.ObservabilityAddon, t string) {
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
