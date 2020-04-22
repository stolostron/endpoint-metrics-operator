#!/bin/bash
echo "BUILD GOES HERE!"

echo "<repo>/<component>:<tag> : $1"

export GOPRIVATE=github.com/open-cluster-management

operator-sdk build $1