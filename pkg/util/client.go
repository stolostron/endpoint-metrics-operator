// Copyright (c) 2021 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project.
package util

import (
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/open-cluster-management/multicluster-monitoring-operator/pkg/apis"
)

const (
	hubKubeConfigPath = "/spoke/hub-kubeconfig/kubeconfig"
)

var (
	log = logf.Log.WithName("util")
)

func CreateHubClient() (client.Client, error) {
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

	// generate the client based off of the config
	hubClient, err := client.New(config, client.Options{Scheme: s})

	if err != nil {
		log.Error(err, "Failed to create hub client")
		return nil, err
	}

	return hubClient, err
}
