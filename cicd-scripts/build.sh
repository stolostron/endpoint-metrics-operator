#!/bin/bash
echo "BUILD GOES HERE!"

echo "<repo>/<component>:<tag> : $1"

git config --global credential.helper store
echo "https://${GITHUB_TOKEN}:x-oauth-basic@github.com" >> ~/.git-credentials
export GOPRIVATE=github.com/open-cluster-management

operator-sdk build $1