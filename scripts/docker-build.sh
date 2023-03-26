#!/usr/bin/env bash

set -eo pipefail

cd "$(dirname "$0")"/..

docker build -f build/package/docker/Dockerfile -t kibernate:latest .

if [[ -n "$(command -v minikube)" ]] && minikube profile list | grep -q kibernate-test; then
  echo "minikube profile kibernate-test exists - importing image into minikube"
  docker save kibernate:latest | (eval "$(minikube docker-env -p kibernate-test)" && docker load)
fi

