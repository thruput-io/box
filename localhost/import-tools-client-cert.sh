#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "This script supports macOS only."
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_ROOT_DIR="${BOX_ROOT:-$(cd "${SCRIPT_DIR}/.." && pwd)}"

P12_PATH="${COMPOSE_ROOT_DIR}/certs/tools-client.p12"
LOGIN_KEYCHAIN="${HOME}/Library/Keychains/login.keychain-db"

if [[ ! -f "${P12_PATH}" ]]; then
  echo "Missing ${P12_PATH}"
  echo "Run: ${SCRIPT_DIR}/generate-tools-client-cert.sh"
  exit 1
fi

echo "Importing client certificate into login keychain..."
security import "${P12_PATH}" \
  -k "${LOGIN_KEYCHAIN}" \
  -P "" \
  -T /usr/bin/curl \
  -T /usr/bin/security \
  >/dev/null

echo "Imported."
echo
echo "You can still test using PEM files (recommended for tooling):"
echo "  curl -vk --cert ${COMPOSE_ROOT_DIR}/certs/tools-client.crt --key ${COMPOSE_ROOT_DIR}/certs/tools-client.key https://tools.web.internal/"
