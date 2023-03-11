#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")"/..

minikube start --driver=docker --kubernetes-version=v1.24.10 -p koldstart-test
minikube profile koldstart-test
function finally() {
  minikube delete
}
trap finally EXIT
eval "$(minikube docker-env)"
docker build --platform linux/amd64 -f build/package/docker/Dockerfile -t koldstart:latest .
kubectl create deployment testtarget --image=ghcr.io/nginxinc/nginx-unprivileged:1.23-alpine --replicas=1 --port=8080
kubectl expose deployment testtarget --port=8080 --target-port=8080
helm install \
  -n default \
  koldstart \
  ./deployments/helm/koldstart \
  --set image.tag=latest \
  --set image.pullPolicy=Never \
  --set service.type=NodePort \
  --set "args[0]=-targetUrl=http://testtarget:8080"
kubectl wait --for=condition=available --timeout=60s deployment/testtarget
kubectl wait --for=condition=available --timeout=60s deployment/koldstart
kubectl run -i --rm test --image=alpine:3 --restart=Never -- /bin/sh -c "set -euo pipefail; apk add curl; curl -sSf 'http://koldstart:8080' | grep 'If you see this page, the nginx web server is successfully installed'"
