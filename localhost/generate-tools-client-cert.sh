#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "This script supports macOS only."
  exit 1
fi

echo "tools client certificates are no longer used."
echo "The tools MCP endpoint is now unprotected (no mTLS)."
exit 0

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_ROOT_DIR="${BOX_ROOT:-$(cd "${SCRIPT_DIR}/.." && pwd)}"

CA_CERT="${COMPOSE_ROOT_DIR}/certs/_ca/rootCA.pem"
CA_KEY="${COMPOSE_ROOT_DIR}/certs/_ca/rootCA-key.pem"
CLIENT_P12="${COMPOSE_ROOT_DIR}/certs/tools-client.p12"
LEGACY_CLIENT_KEY="${COMPOSE_ROOT_DIR}/certs/tools-client.key"
LEGACY_CLIENT_CERT="${COMPOSE_ROOT_DIR}/certs/tools-client.crt"
TMP_DIR="${COMPOSE_ROOT_DIR}/certs/.tmp"
CLIENT_KEY="${TMP_DIR}/tools-client.key"
CLIENT_CSR="${TMP_DIR}/tools-client.csr"
CLIENT_CERT="${TMP_DIR}/tools-client.crt"
EXTFILE="${TMP_DIR}/tools-client.ext"

if [[ ! -f "${CA_CERT}" || ! -f "${CA_KEY}" ]]; then
  echo "Missing CA materials under ${COMPOSE_ROOT_DIR}/certs/_ca"
  echo "Run: make -C ${COMPOSE_ROOT_DIR} generate-certs"
  exit 1
fi

# Keep the output deterministic: only a PKCS#12 bundle is persisted.
# If legacy PEM outputs exist from older versions, remove them.
rm -f "${LEGACY_CLIENT_KEY}" "${LEGACY_CLIENT_CERT}" 2>/dev/null || true

if [[ -f "${CLIENT_P12}" ]]; then
  echo "Client PKCS#12 bundle already exists:"
  ls -la "${CLIENT_P12}" 2>/dev/null || true
  echo "Delete it if you want to re-issue."
  exit 0
fi

mkdir -p "${TMP_DIR}"

cat >"${EXTFILE}" <<'EOF'
basicConstraints=CA:FALSE
keyUsage=digitalSignature
extendedKeyUsage=clientAuth
EOF

echo "Generating client key..."
openssl genrsa -out "${CLIENT_KEY}" 2048

echo "Generating CSR..."
openssl req -new -key "${CLIENT_KEY}" -out "${CLIENT_CSR}" -subj "/CN=tools-client"

echo "Signing client certificate..."
openssl x509 -req \
  -in "${CLIENT_CSR}" \
  -CA "${CA_CERT}" -CAkey "${CA_KEY}" -CAcreateserial \
  -out "${CLIENT_CERT}" \
  -days 365 \
  -extfile "${EXTFILE}"

echo "Creating PKCS#12 bundle for Keychain import..."
openssl pkcs12 -export \
  -out "${CLIENT_P12}" \
  -inkey "${CLIENT_KEY}" \
  -in "${CLIENT_CERT}" \
  -certfile "${COMPOSE_ROOT_DIR}/certs/dev-root-ca.crt" \
  -passout pass:""

rm -rf "${TMP_DIR}"

echo "Created:"
ls -la "${CLIENT_P12}"

echo
echo "Test (should FAIL without client cert):"
echo "  curl -vk https://tools.web.internal/"
echo
echo "Test (should SUCCEED with client cert):"
echo "  curl -vk --cert-type P12 --cert ${CLIENT_P12}: https://tools.web.internal/"
