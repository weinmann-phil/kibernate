#!/usr/bin/env bash

set -eo pipefail

cd "$(dirname "$0")"/..

BRANCH_NAME=$(git rev-parse --abbrev-ref HEAD)
rm -fr kibernate-helm
git clone -b $BRANCH_NAME --single-branch --depth 1 https://github.com/kibernate/kibernate-helm.git
rm -fr kibernate-helm/.git
