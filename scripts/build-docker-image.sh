#!/usr/bin/env bash

set -eo pipefail

cd "$(dirname "$0")"/..
eval "$(minikube -p kibernate-test docker-env)"
docker build -f build/package/docker/Dockerfile -t kibernate:latest .