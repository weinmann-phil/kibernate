#!/usr/bin/env bash

set -eo pipefail

cd "$(dirname "$0")"/..

helm upgrade --install \
  -n default \
  kibernate \
  ./deployments/helm/kibernate \
  --set image.repository=kibernate \
  --set image.tag=latest \
  --set image.pullPolicy=Never \
  --set kibernate.service=testtarget \
  --set kibernate.namespace=default \
  --set kibernate.defaultWaitType=connect \
  --set kibernate.idleTimeoutSecs=60 \
  --set kibernate.deployment=testtarget