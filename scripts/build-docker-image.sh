#!/usr/bin/env bash

set -eo pipefail

cd "$(dirname "$0")"/..
docker build -f build/package/docker/Dockerfile -t kibernate:latest .