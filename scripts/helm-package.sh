#!/usr/bin/env bash

set -eo pipefail

cd "$(dirname "$0")"/../deployments/helm
helm lint --strict kibernate
helm package kibernate --destination .
