#!/usr/bin/env bash

set -eo pipefail

if [[ -z "$1" ]]; then
  echo "Usage: $0 <platforms>"
  exit 1
fi

cd "$(dirname "$0")"/..

if [[ -n "$(command -v minikube)" ]] && minikube profile list | grep -q kibernate-test; then
  echo "minikube profile kibernate-test exists - assuming it is build environment"
  eval "$(minikube -p kibernate-test docker-env)"
fi

docker buildx build --platform="$1" -f build/package/docker/Dockerfile -t kibernate:latest .