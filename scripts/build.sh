#!/usr/bin/env bash

set -e
set -o pipefail
set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Building nuon-ext-render..."
cd "$ROOT_DIR"
GOWORK=off go build -ldflags "-s -w" -o "$ROOT_DIR/nuon-ext-render" .

echo "Built: $ROOT_DIR/nuon-ext-render"
