#!/usr/bin/env bash

set -eo pipefail

if [[ -z "$1" ]]; then
  echo "Usage: $0 <platforms>"
  exit 1
fi

cd "$(dirname "$0")"/..

docker buildx build --output type=registry --platform="$1" -f build/package/docker/Dockerfile -t kibernate:latest .

if [[ -n "$(command -v minikube)" ]] && minikube profile list | grep -q kibernate-test; then
  echo "minikube profile kibernate-test exists - importing image into minikube"
  docker save kibernate:latest | (eval "$(minikube docker-env -p kibernate-test)" && docker load)
fi

