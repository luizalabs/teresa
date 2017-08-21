TERESA_IMAGE_NAME ?= teresa
TERESA_IMAGE_VERSION ?= latest
TERESA_K8S_CONFIG_FILE ?= ~/.kube/config
DOCKER_K8S_CONFIG_FILE = /config
DOCKER_SECRETS_PRIVATE_KEY = /teresa.rsa
DOCKER_SECRETS_PUBLIC_KEY = /teresa.rsa.pub
IMAGE_INSTANCE = default
TERESA_DOCKER_PORT ?= 50051
BUILD_VERSION ?= $(shell git describe --always --tags)
BUILD_HOME = github.com/luizalabs/teresa

DOCKER_RUN_CMD=docker run \
	-e TERESA_SECRETS_PRIVATE_KEY=$(DOCKER_SECRETS_PRIVATE_KEY) \
	-e TERESA_SECRETS_PUBLIC_KEY=$(DOCKER_SECRETS_PUBLIC_KEY) \
	-e TERESA_STORAGE_AWS_KEY=$(TERESA_STORAGE_AWS_KEY) \
	-e TERESA_STORAGE_AWS_SECRET=$(TERESA_STORAGE_AWS_SECRET) \
	-e TERESA_STORAGE_AWS_REGION=$(TERESA_STORAGE_AWS_REGION) \
	-e TERESA_STORAGE_AWS_BUCKET=$(TERESA_STORAGE_AWS_BUCKET) \
	-e TERESA_K8S_CONFIG_FILE=$(DOCKER_K8S_CONFIG_FILE) \
	-v $(TERESA_K8S_CONFIG_FILE):$(DOCKER_K8S_CONFIG_FILE) \
	-v $(TERESA_SECRETS_PRIVATE_KEY):$(DOCKER_SECRETS_PRIVATE_KEY) \
	-v $(TERESA_SECRETS_PUBLIC_KEY):$(DOCKER_SECRETS_PUBLIC_KEY) \
	-p $(TERESA_DOCKER_PORT):50051

help:
	@echo "Targets are:\n"
	@echo "build"
	@echo " build the teresa server docker image"
	@echo
	@echo "run"
	@echo " run the teresa server docker image"
	@echo
	@echo "start"
	@echo " run the teresa server docker image as a daemon"
	@echo
	@echo "stop"
	@echo " stop the teresa server docker image"
	@echo
	@echo "shell"
	@echo " run a bash shell on the docker image"
	@echo
	@echo "run-server"
	@echo " run the teresa server locally"
	@echo
	@echo "build-client"
	@echo " build the teresa client 'teresa'"
	@echo
	@echo "build-server"
	@echo " build the teresa server 'teresa-server'"
	@echo
	@echo "gen-grpc-stubs"
	@echo " generate grpc code, only used for development"
	@echo
	@echo "To run the container or server you'll have to set the following env variables:"
	@echo
	@echo "	TERESA_STORAGE_AWS_KEY"
	@echo "	TERESA_STORAGE_AWS_SECRET"
	@echo "	TERESA_STORAGE_AWS_REGION"
	@echo "	TERESA_STORAGE_AWS_BUCKET"
	@echo "	TERESA_SECRETS_PUBLIC_KEY"
	@echo "	TERESA_SECRETS_PRIVATE_KEY"
	@echo "	TERESA_K8S_CONFIG_FILE - optional, default to ~/.kube/config"
	@echo "	TERESA_DOCKER_PORT     - optional, defaults to 50051"
	@echo "	TERESA_IMAGE_NAME      - optional, defaults to teresa"
	@echo "	TERESA_IMAGE_VERSION   - optional, defaults to latest"
	@echo
	@echo "To build the server docker image the following env variables are used:"
	@echo
	@echo "	TERESA_IMAGE_NAME      - optional, defaults to teresa"
	@echo "	TERESA_IMAGE_VERSION   - optional, defaults to latest"
	@echo "	BUILD_VERSION          - optional, defaults to git tag"

all: help

build:
	@docker build -t $(TERESA_IMAGE_NAME):$(TERESA_IMAGE_VERSION) .

run:
	@$(DOCKER_RUN_CMD) --rm --name $(TERESA_IMAGE_NAME)-$(IMAGE_INSTANCE) $(TERESA_IMAGE_NAME):$(TERESA_IMAGE_VERSION)

start:
	@$(DOCKER_RUN_CMD) -d $(TERESA_IMAGE_NAME):$(TERESA_IMAGE_VERSION)

stop:
	@docker stop $(TERESA_IMAGE_NAME)-$(IMAGE_INSTANCE)

shell:
	@docker run --rm -it --name $(TERESA_IMAGE_NAME)-$(IMAGE_INSTANCE) $(TERESA_IMAGE_NAME):$(TERESA_IMAGE_VERSION) /bin/bash

run-server:
	@go run ./cmd/server/main.go run

build-server:
	@go build -ldflags "-X $(BUILD_HOME)/pkg/version.Version=$(BUILD_VERSION)" -o teresa-server $(BUILD_HOME)/cmd/server

build-client:
	@go build -ldflags "-X $(BUILD_HOME)/pkg/version.Version=$(BUILD_VERSION)" -o teresa $(BUILD_HOME)/cmd/client

gen-grpc-stubs:
	@protoc --go_out=plugins=grpc:. ./pkg/protobuf/user/*.proto
	@protoc --go_out=plugins=grpc:. ./pkg/protobuf/team/*.proto
	@protoc --go_out=plugins=grpc:. ./pkg/protobuf/app/*.proto
	@protoc --go_out=plugins=grpc:. ./pkg/protobuf/deploy/*.proto

helm-lint:
	@helm lint helm/chart/teresa

update-chart: helm-lint
	@helm package helm/chart/teresa
	@mkdir repo
	@mv teresa-*.tgz repo
	@helm repo index repo --url http://helm.k8s.magazineluiza.com
	@aws s3 sync repo s3://helm.k8s.magazineluiza.com --delete
	@rm -rf repo
