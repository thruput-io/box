#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "This script supports macOS only."
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

KEYCHAIN_PATH="${HOME}/Library/Keychains/infra-localhost.keychain-db"
CERT_PATH="${COMPOSE_ROOT_DIR}/certs/dev-root-ca.crt"
RESOLVER_DIR="/etc/resolver"
RESOLVER_FILE="${RESOLVER_DIR}/internal"
COREDNS_HOST="127.0.0.1"
COREDNS_PORT="5354"

if [[ ! -f "${CERT_PATH}" ]]; then
  echo "Missing certificate: ${CERT_PATH}"
  echo "Generate certificates first (for example: make -C ${COMPOSE_ROOT_DIR} generate-certs)."
  exit 1
fi

if [[ ! -f "${KEYCHAIN_PATH}" ]]; then
  security create-keychain -p "" "${KEYCHAIN_PATH}"
fi

security unlock-keychain -p "" "${KEYCHAIN_PATH}"

CURRENT_KEYCHAINS="$(security list-keychains -d user | tr -d '"')"
if ! printf '%s\n' "${CURRENT_KEYCHAINS}" | grep -Fq "${KEYCHAIN_PATH}"; then
  # shellcheck disable=SC2086
  security list-keychains -d user -s "${KEYCHAIN_PATH}" ${CURRENT_KEYCHAINS}
fi

CERT_SHA1="$(openssl x509 -in "${CERT_PATH}" -noout -fingerprint -sha1 | awk -F= '{print $2}' | tr -d ':')"
if ! security find-certificate -a -Z "${KEYCHAIN_PATH}" | tr -d ' ' | grep -Fq "${CERT_SHA1}"; then
  security add-trusted-cert -d -r trustRoot -p ssl -k "${KEYCHAIN_PATH}" "${CERT_PATH}"
fi

sudo mkdir -p "${RESOLVER_DIR}"
printf 'nameserver %s\nport %s\n' "${COREDNS_HOST}" "${COREDNS_PORT}" | sudo tee "${RESOLVER_FILE}" >/dev/null

sudo dscacheutil -flushcache || true
sudo killall -HUP mDNSResponder || true

echo "Configured host DNS for *.internal via ${COREDNS_HOST}:${COREDNS_PORT}."
echo "Trusted cert imported into ${KEYCHAIN_PATH}."

# Import root CA into Firefox NSS certificate stores (Firefox ignores the macOS keychain)
CERT_NICKNAME="infra-localhost-root-ca"
if command -v certutil >/dev/null 2>&1; then
  for nss_db in "${HOME}/Library/Application Support/Mozilla/Firefox/Profiles"/*/; do
    if [[ -f "${nss_db}cert9.db" ]]; then
      certutil -D -n "${CERT_NICKNAME}" -d "sql:${nss_db}" 2>/dev/null || true
      certutil -A -n "${CERT_NICKNAME}" -t "CT,C,C" -i "${CERT_PATH}" -d "sql:${nss_db}"
      echo "Imported root CA into Firefox profile: ${nss_db}"
    fi
  done
else
  echo "Warning: certutil not found. Install nss (brew install nss) to trust the root CA in Firefox."
fi