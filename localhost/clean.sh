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

# Remove the root CA from the login keychain (added for Chrome)
LOGIN_KEYCHAIN="${HOME}/Library/Keychains/login.keychain-db"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
CERT_PATH="${COMPOSE_ROOT_DIR}/certs/dev-root-ca.crt"
if [[ -f "${LOGIN_KEYCHAIN}" && -f "${CERT_PATH}" ]]; then
  CERT_SHA1="$(openssl x509 -in "${CERT_PATH}" -noout -fingerprint -sha1 | awk -F= '{print $2}' | tr -d ':')"
  if security find-certificate -a -Z "${LOGIN_KEYCHAIN}" | tr -d ' ' | grep -Fq "${CERT_SHA1}"; then
    security remove-trusted-cert -d "${CERT_PATH}" 2>/dev/null || true
    CERT_CN="$(openssl x509 -in "${CERT_PATH}" -noout -subject -nameopt multiline | awk -F= '/commonName/{gsub(/^ +/,"",$2); print $2}')"
    if [[ -n "${CERT_CN}" ]]; then
      security delete-certificate -c "${CERT_CN}" "${LOGIN_KEYCHAIN}" 2>/dev/null || true
    fi
    echo "Removed root CA trust from login keychain."
  fi
fi

# Clear Chrome's certificate verification cache
CHROME_CERT_CACHE="${HOME}/Library/Application Support/Google/Chrome/CertificateVerification"
if [[ -d "${CHROME_CERT_CACHE}" ]]; then
  rm -rf "${CHROME_CERT_CACHE}"
  echo "Cleared Chrome certificate verification cache."
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