#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_ROOT_DIR="${BOX_ROOT:-$(cd "${SCRIPT_DIR}/.." && pwd)}"

TOKEN_PATH="${COMPOSE_ROOT_DIR}/certs/tools-token.txt"

mkdir -p "$(dirname "${TOKEN_PATH}")"

if [[ -f "${TOKEN_PATH}" ]] && [[ -s "${TOKEN_PATH}" ]]; then
  exit 0
fi

umask 077

if command -v openssl >/dev/null 2>&1; then
  # ~64 chars base64url-ish, plenty for a shared secret.
  openssl rand -base64 48 | tr -d '\n' >"${TOKEN_PATH}"
else
  # Fallback: not as strong, but avoids a hard dependency.
  LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 64 >"${TOKEN_PATH}"
fi

chmod 600 "${TOKEN_PATH}" 2>/dev/null || true

echo "Generated tools token at: ${TOKEN_PATH}"