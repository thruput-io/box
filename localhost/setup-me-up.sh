#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "This script supports macOS only."
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}" || echo "${BASH_SOURCE[0]}")")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# --- Install the box command (sets BOX_ROOT) ---
echo "=== Installing box command ==="
"${SCRIPT_DIR}/install-box.sh" "${PROJECT_ROOT}"
export BOX_ROOT="${PROJECT_ROOT}"

# --- Install Homebrew dependencies ---
echo ""
echo "=== Installing dependencies ==="
if ! command -v brew >/dev/null 2>&1; then
  echo "Homebrew is not installed. Install it from https://brew.sh"
  exit 1
fi
brew bundle --file="${BOX_ROOT}/localhost/Brewfile"

# --- Install PowerShell via dotnet tool ---
echo ""
echo "=== Installing PowerShell ==="
if ! command -v pwsh >/dev/null 2>&1; then
  dotnet tool install --global PowerShell
else
  echo "PowerShell (pwsh) is already installed."
fi

echo ""
echo "=== All done ==="
echo "You're ready to go. Run 'make up' to start the environment."
