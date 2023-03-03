MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))

GO ?= go

REGISTRY ?= quay.io/rhoas
IMAGE_NAME ?= observability-remote-write-proxy
VERSION ?= "$$(git rev-parse --short=7 HEAD)"

BUILD_BINARY_NAME ?= observability-remote-write-proxy

build:
	${GO} build -o ${PROJECT_PATH}/${BUILD_BINARY_NAME} ${PROJECT_PATH}

docker/login:
	docker --config=${DOCKER_CONFIG} login -u ${QUAY_USER} -p ${QUAY_TOKEN} quay.io
.PHONY: docker/login

image/build:
	docker build -t ${REGISTRY}/${IMAGE_NAME}:${VERSION} -f ./Dockerfile .
.PHONY: image/build

image/push: image/build
	docker --config=${DOCKER_CONFIG} push ${REGISTRY}/${IMAGE_NAME}:${VERSION}
.PHONY: image/push
