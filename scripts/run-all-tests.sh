#!/usr/bin/env bash

set -eo pipefail

cd "$(dirname "$0")"/..
./scripts/prepare-testing-env.sh
function finally() {
  ./scripts/tear-down-testing-env.sh
}
trap finally EXIT
./scripts/build-docker-image.sh
kubectl create deployment testtarget --image=ghcr.io/nginxinc/nginx-unprivileged:1.23-alpine --replicas=1 --port=8080
kubectl expose deployment testtarget --port=8080 --target-port=8080
./scripts/install-helm-chart.sh
kubectl wait --for=condition=available --timeout=60s deployment/testtarget
kubectl wait --for=condition=available --timeout=60s deployment/kibernate
kubectl run -i --rm test --image=alpine:3 --restart=Never -- /bin/sh -c "set -eo pipefail; apk add curl; curl -s 'http://kibernate:8080' | grep 'Thank you for using nginx.'"
