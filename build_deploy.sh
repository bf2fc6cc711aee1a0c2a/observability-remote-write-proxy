#!/bin/bash -ex
#
# Copyright (c) 2018 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# This script builds and deploys the Managed Kafka Service Observability Config. In order to
# work, it needs the following variables defined in the CI/CD configuration of
# the project:
#
# QUAY_USER - The name of the robot account used to push images to
# 'quay.io', for example 'openshift-unified-hybrid-cloud+jenkins'.
#
# QUAY_TOKEN - The token of the robot account used to push images to
# 'quay.io'.
#
# The machines that run this script need to have access to internet, so that
# the built images can be pushed to quay.io.

# Set the variable required to login and push images to the registry
QUAY_USER=${QUAY_USER_NAME:-$RHOAS_QUAY_USER}
QUAY_TOKEN=${QUAY_USER_PASSWORD:-$RHOAS_QUAY_TOKEN}

DOCKER_CONFIG="${PWD}/.docker"

# Set up the docker config directory
mkdir -p "${DOCKER_CONFIG}"

# Login to docker
make docker/login

# Push the image to the registry, Makefile target includes the build
make image/push
