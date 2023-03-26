#!/usr/bin/env bash

set -eo pipefail

cd "$(dirname "$0")"/..

helm upgrade --install \
  -n default \
  kibernate \
  ./deployments/helm/kibernate \
  --set image.tag=latest \
  --set image.pullPolicy=Never \
  --set "args[0]=-service=testtarget" \
  --set "args[1]=-servicePort=8080" \
  --set "args[2]=-deployment=testtarget" \
  --set "args[3]=-namespace=default" \
  --set "args[4]=-defaultWaitType=connect" \
  --set "args[5]=-idleTimeoutSecs=60"