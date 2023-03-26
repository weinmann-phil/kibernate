#!/usr/bin/env bash

set -eo pipefail

if [[ -z "$1" ]]; then
  echo "Usage: $0 <tags>"
  exit 1
fi

cd "$(dirname "$0")"/..

for tag in "$@"; do
  tags="$tags -t ghcr.io/kibernate/kibernate:$tag"
done

docker buildx build --platforms linux/amd64 $tags --push -f build/package/docker/Dockerfile .
