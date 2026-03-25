#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "This script supports macOS only."
  exit 1
fi

KEYCHAIN_PATH="${HOME}/Library/Keychains/infra-localhost.keychain-db"
RESOLVER_FILE="/etc/resolver/internal"

if [[ -f "${RESOLVER_FILE}" ]]; then
  sudo rm -f "${RESOLVER_FILE}"
fi

if [[ -f "${KEYCHAIN_PATH}" ]]; then
  security delete-keychain "${KEYCHAIN_PATH}"
fi

sudo dscacheutil -flushcache || true
sudo killall -HUP mDNSResponder || true

echo "Removed host DNS override for *.internal."
echo "Deleted keychain ${KEYCHAIN_PATH}."

# Remove root CA from Firefox NSS certificate stores
CERT_NICKNAME="infra-localhost-root-ca"
if command -v certutil >/dev/null 2>&1; then
  for nss_db in "${HOME}/Library/Application Support/Mozilla/Firefox/Profiles"/*/; do
    if [[ -f "${nss_db}cert9.db" ]]; then
      certutil -D -n "${CERT_NICKNAME}" -d "sql:${nss_db}" 2>/dev/null || true
      echo "Removed root CA from Firefox profile: ${nss_db}"
    fi
  done
fi