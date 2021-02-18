// Copyright (c) 2021 Red Hat, Inc.

package util

import (
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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

	// generate the client based on the config
	hubClient, err := client.New(config, client.Options{})
	if err != nil {
		log.Error(err, "Failed to create hub client")
		return nil, err
	}

	return hubClient, err
}
