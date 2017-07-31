TERESA_API_NAME=Teresa
IMAGE_NAME=teresa
IMAGE_VERSION=0.1.1
IMAGE_INSTANCE=default
TERESA_DOCKER_PORT ?= 8080
TERESA_VERSION ?= $(shell git describe --always --tags)
TERESA_API=github.com/luizalabs/teresa-api

DOCKER_RUN_CMD=docker run \
	-e TERESAK8S_HOST=$(TERESAK8S_HOST) \
	-e TERESAK8S_USERNAME=$(TERESAK8S_USERNAME) \
	-e TERESAK8S_PASSWORD=$(TERESAK8S_PASSWORD) \
	-e TERESAK8S_INSECURE=$(TERESAK8S_INSECURE) \
	-e TERESAFILESTORAGE_TYPE=$(TERESAFILESTORAGE_TYPE) \
	-e TERESAFILESTORAGE_AWS_KEY=$(TERESAFILESTORAGE_AWS_KEY) \
	-e TERESAFILESTORAGE_AWS_SECRET=$(TERESAFILESTORAGE_AWS_SECRET) \
	-e TERESAFILESTORAGE_AWS_REGION=$(TERESAFILESTORAGE_AWS_REGION) \
	-e TERESAFILESTORAGE_AWS_BUCKET=$(TERESAFILESTORAGE_AWS_BUCKET) \
	-p $(TERESA_DOCKER_PORT):8080

help:
	@echo "Targets are:\n"
	@echo "build"
	@echo " build the teresa API server docker image"
	@echo
	@echo "run"
	@echo " run the teresa API docker image"
	@echo
	@echo "start"
	@echo " run the teresa API docker image as a daemon"
	@echo
	@echo "stop"
	@echo " stop the teresa API docker image"
	@echo
	@echo "To run the API container you'll have to set the following env variables:"
	@echo
	@echo "	TERESAK8S_HOST"
	@echo "	TERESAK8S_USERNAME"
	@echo "	TERESAK8S_PASSWORD"
	@echo "	TERESAK8S_INSECURE"
	@echo "	TERESAFILESTORAGE_TYPE"
	@echo "	TERESAFILESTORAGE_AWS_KEY"
	@echo "	TERESAFILESTORAGE_AWS_SECRET"
	@echo "	TERESAFILESTORAGE_AWS_REGION"
	@echo "	TERESAFILESTORAGE_AWS_BUCKET"
	@echo "	TERESA_DOCKER_PORT - optional, defaults to 8080"
	@echo

all: help

build:
	docker build -t $(IMAGE_NAME):$(IMAGE_VERSION) .

run:
	$(DOCKER_RUN_CMD) --rm --name $(IMAGE_NAME)-$(IMAGE_INSTANCE) $(IMAGE_NAME):$(IMAGE_VERSION)

start:
	$(DOCKER_RUN_CMD) -d $(IMAGE_NAME):$(IMAGE_VERSION)

stop:
	docker stop $(IMAGE_NAME)-$(IMAGE_INSTANCE)

shell:
	docker run --rm --name $(IMAGE_NAME)-$(IMAGE_INSTANCE) -i -t $(IMAGE_NAME):$(IMAGE_VERSION) /bin/bash

run-api-server:
	go run ./cmd/server/main.go run --port 8080

build-server:
	@go build -ldflags "-X $(TERESA_API)/pkg/version.Version=$(TERESA_VERSION)" -o teresa-server $(TERESA_API)/cmd/server

build-client:
	@go build -ldflags "-X $(TERESA_API)/pkg/version.Version=$(TERESA_VERSION)" -o teresa $(TERESA_API)/cmd/client

gen-grpc-stubs:
	@protoc --go_out=plugins=grpc:. ./pkg/protobuf/user/*.proto
	@protoc --go_out=plugins=grpc:. ./pkg/protobuf/team/*.proto
	@protoc --go_out=plugins=grpc:. ./pkg/protobuf/app/*.proto
	@protoc --go_out=plugins=grpc:. ./pkg/protobuf/deploy/*.proto
