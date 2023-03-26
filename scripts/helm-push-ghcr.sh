#!/usr/bin/env bash

set -eo pipefail

cd "$(dirname "$0")"/../deployments/helm
CHART_VERSION=$(grep -oe 'version: "\?[^"]*"\?' kibernate/Chart.yaml | cut -d' ' -f2 | tr -d '"')
APPLICATION_VERSION=$(grep -oe 'appVersion: "\?[^"]*"\?' kibernate/Chart.yaml | cut -d' ' -f2 | tr -d '"')
echo "CHART_VERSION=$CHART_VERSION"
echo "APPLICATION_VERSION=$APPLICATION_VERSION"
# ensure docker image ghcr.io/kibernate/kibernate:$APPLICATION_VERSION does exist:
docker manifest inspect ghcr.io/kibernate/kibernate:"$APPLICATION_VERSION" > /dev/null
helm push kibernate-"$CHART_VERSION".tgz oci://ghcr.io/kibernate
