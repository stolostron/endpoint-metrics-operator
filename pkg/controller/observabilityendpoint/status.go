// Copyright (c) 2020 Red Hat, Inc.

package observabilityendpoint

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	oav1beta1 "github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis/observability/v1beta1"
)

var (
	conditions = map[string]map[string]string{
		"Ready": map[string]string{
			"reason":  "Deployed",
			"message": "Metrics collector deployed and functional"},
		"Disabled": map[string]string{
			"reason":  "Disabled",
			"message": "enableMetrics is set to False"},
		"NotSupported": map[string]string{
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
