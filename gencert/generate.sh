#!/bin/sh
set -euo pipefail

required_sans='DNS:*.web.internal DNS:*.mock.internal DNS:*.internal DNS:*.fe-dev.internal DNS:login.microsoftonline.com DNS:login.microsoft.com DNS:*.servicebus.windows.net DNS:*.postgres.database.azure.com DNS:wiremock.local DNS:localhost'

TLS_CERT_PATH="/certs/tls-cert.pem"
TLS_KEY_PATH="/certs/tls-key.pem"
CA_CERT_PATH="/certs/dev-root-ca.crt"
IDENTITY_KEY_PATH="/certs/identity-signing.key"

TOOLS_CLIENT_P12_PATH="/certs/tools-client.p12"

has_required_sans() {
    cert_file="$1"

    if ! cert_text="$(openssl x509 -in "$cert_file" -noout -text 2>/dev/null)"; then
        return 1
    fi

    for san in $required_sans; do
        if ! printf '%s' "$cert_text" | grep -F "$san" >/dev/null; then
            return 1
        fi
    done

    return 0
}

regen_tls="1"

# Check if certificates already exist and contain all required SANs.
if [ -f "$TLS_CERT_PATH" ] && \
   [ -f "$TLS_KEY_PATH" ] && \
   [ -f "$CA_CERT_PATH" ] && \
   [ -f "$IDENTITY_KEY_PATH" ] && \
   has_required_sans "$TLS_CERT_PATH"; then
  echo "✓ Certificates already exist with required SANs. Skipping TLS/CA regeneration."
  regen_tls="0"
else
  echo "Generating certificates..."
fi

# Define project-local CAROOT to keep the CA self-contained in the project.
export CAROOT="/certs/_ca"
mkdir -p "$CAROOT"

# Helper function to run mkcert
run_mkcert() {
    /usr/local/bin/mkcert "$@"
}

mkdir -p /certs

if [ "$regen_tls" = "1" ]; then
  rm -f "$TLS_CERT_PATH" "$TLS_KEY_PATH" "$CA_CERT_PATH" "$IDENTITY_KEY_PATH"

  echo "Generating wildcard certificate for *.internal domains and mock services..."
  # mkcert will automatically create the CA in $CAROOT if it doesn't exist.
  run_mkcert -key-file "$TLS_KEY_PATH" -cert-file "$TLS_CERT_PATH" \
      "*.web.internal" \
      "*.mock.internal" \
      "*.internal" \
      "*.fe-dev.internal" \
      "login.microsoftonline.com" \
      "login.microsoft.com" \
      "*.servicebus.windows.net" \
      "*.postgres.database.azure.com" \
      "wiremock.local" \
      localhost

  echo "Copying CA certificate..."
  cp "$CAROOT/rootCA.pem" "$CA_CERT_PATH"

  echo "Generating RSA private key for JWT signing..."
  openssl genrsa -out "$IDENTITY_KEY_PATH" 2048 2>/dev/null
fi

rm -f "$TOOLS_CLIENT_P12_PATH" 2>/dev/null || true


echo "✓ Certificate generation complete!"
echo ""
echo "Certificates created in /certs directory:"
echo "  - tls-cert.pem, tls-key.pem (TLS wildcard cert)"
echo "  - dev-root-ca.crt (root CA for test container)"
echo "  - identity-signing.key (RSA key for JWT signing)"
echo "  - _ca/ (local CA root directory)"
