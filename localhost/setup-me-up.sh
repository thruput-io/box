#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "This script supports macOS only."
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# --- Install Homebrew dependencies ---
echo "=== Installing dependencies ==="
if ! command -v brew >/dev/null 2>&1; then
  echo "Homebrew is not installed. Install it from https://brew.sh"
  exit 1
fi
brew bundle --file="${SCRIPT_DIR}/Brewfile"

# --- Install the box command ---
echo ""
echo "=== Installing box command ==="
"${SCRIPT_DIR}/install-box.sh" "${PROJECT_ROOT}"

echo ""
echo "=== All done ==="
echo "You're ready to go. Run 'make up' to start the environment."
