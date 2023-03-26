#!/usr/bin/env bash

set -eo pipefail

if [[ -z "$1" ]]; then
  echo "Usage: $0 <platforms>"
  exit 1
fi

cd "$(dirname "$0")"/..
docker buildx build --platform="$1" -f build/package/docker/Dockerfile -t kibernate:latest .