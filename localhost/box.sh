#!/usr/bin/env bash
set -euo pipefail

if [ -z "${BOX_ROOT:-}" ]; then
  echo "BOX_ROOT is not set. Run 'make install-box' from Hosting.Compose."
  exit 1
fi

if [ ! -d "$BOX_ROOT" ]; then
  echo "Configured BOX_ROOT does not exist: $BOX_ROOT"
  exit 1
fi

exec make -C "$BOX_ROOT" "$@"
