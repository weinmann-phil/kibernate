#!/usr/bin/env bash

set -eo pipefail

if [[ -z "$1" ]]; then
  echo "Usage: $0 <tags>"
  exit 1
fi

for tag in "$@"; do
  docker tag kibernate:latest "ghcr.io/kibernate/kibernate:$tag"
  docker push "ghcr.io/kibernate/kibernate:$tag"
done
