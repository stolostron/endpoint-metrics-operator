// Copyright (c) 2021 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project.
package controller

import (
	"github.com/open-cluster-management/endpoint-metrics-operator/pkg/controller/status"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, status.Add)
}
