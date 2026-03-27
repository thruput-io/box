#!/bin/sh
set -eu

DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
PORTAL_DIR="$(dirname -- "$DIR")"

mkdir -p "$PORTAL_DIR/data"

node "$DIR/runtime.js" > "$PORTAL_DIR/data/runtime.json"
