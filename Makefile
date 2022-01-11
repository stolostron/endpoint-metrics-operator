# Copyright (c) 2021 Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project.
include ./cicd-scripts/Configfile

-include $(shell [ -f ".build-harness-bootstrap" ] || curl -sL -o .build-harness-bootstrap -H "Authorization: token $(GITHUB_TOKEN)" -H "Accept: application/vnd.github.v3.raw" "https://raw.github.com/stolostron/build-harness-extensions/main/templates/Makefile.build-harness-bootstrap"; echo .build-harness-bootstrap)

docker-binary:
	CGO_ENABLED=0 go build -a -installsuffix cgo -o build/_output/bin/endpoint-monitoring-operator main.go

copyright-check:
	./cicd-scripts/copyright-check.sh $(TRAVIS_BRANCH)

unit-tests:
	@echo "TODO: Run unit-tests"
	go test ./... -v -coverprofile cover.out
	go tool cover -html=cover.out -o=cover.html

e2e-tests:
	@echo "TODO: Run e2e-tests"
