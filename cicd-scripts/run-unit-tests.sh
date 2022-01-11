#!/bin/bash
# Copyright (c) 2021 Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project.
echo "<repo>/<component>:<tag> : $1"

go test ./...
