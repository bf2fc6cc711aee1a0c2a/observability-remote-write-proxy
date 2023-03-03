REGISTRY ?= quay.io/rhoas
IMAGE_NAME ?= observability-remote-write-proxy
VERSION ?= "$$(git rev-parse --short=7 HEAD)"

docker/login:
	docker --config=${DOCKER_CONFIG} login -u ${QUAY_USER} -p ${QUAY_TOKEN} quay.io
.PHONY: docker/login

image/build:
	docker build -t ${REGISTRY}/${IMAGE_NAME}:${VERSION} -f ./Dockerfile .
.PHONY: image/build

image/push: image/build
	docker --config=${DOCKER_CONFIG} push ${REGISTRY}/${IMAGE_NAME}:${VERSION}
.PHONY: image/push
