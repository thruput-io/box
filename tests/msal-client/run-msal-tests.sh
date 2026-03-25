#!/usr/bin/env bash
set -euo pipefail

CERT_SOURCE="${MSAL_CA_CERT_PATH:-/certs/dev-root-ca.crt}"
CERT_TARGET="/usr/local/share/ca-certificates/mkcert.crt"

if [ -f "$CERT_SOURCE" ]; then
  cp "$CERT_SOURCE" "$CERT_TARGET"
fi

update-ca-certificates >/dev/null 2>&1 || true

mkdir -p /root/.pki/nssdb
if [ ! -f /root/.pki/nssdb/cert9.db ]; then
  certutil -d sql:/root/.pki/nssdb -N --empty-password
fi

certutil -d sql:/root/.pki/nssdb -D -n "it-local-ca" >/dev/null 2>&1 || true
certutil -d sql:/root/.pki/nssdb -A -t "CT,C,C" -n "it-local-ca" -i "$CERT_TARGET"

exec npm test