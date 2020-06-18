#!/bin/bash
# Copyright (c) 2020 Red Hat, Inc.

set -e

make docker-binary

echo "Building endpoint-monitoring-operator image"
export DOCKER_IMAGE_AND_TAG=${1}
export DOCKER_FILE=Dockerfile
make docker/build