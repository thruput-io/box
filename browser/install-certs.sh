#!/bin/sh
set -e
set -u

CA_CERT="/certs/dev-root-ca.crt"

if [ ! -f "${CA_CERT}" ]; then
    echo "No custom CA certificate found at ${CA_CERT}, skipping."
    exit 0
fi

# Install into system trust store
echo "Installing custom CA into system trust store..."
mkdir -p /usr/local/share/ca-certificates
cp "${CA_CERT}" /usr/local/share/ca-certificates/dev-root-ca.crt
update-ca-certificates

# Install into Firefox NSS database
PROFILE_DIR="/config/profile"
mkdir -p "${PROFILE_DIR}"

if command -v certutil >/dev/null 2>&1; then
    echo "Importing custom CA into Firefox NSS database..."
    certutil -D -n "dev-root-ca" -d "sql:${PROFILE_DIR}" 2>/dev/null || true
    certutil -A -n "dev-root-ca" -t "CT,C,C" -i "${CA_CERT}" -d "sql:${PROFILE_DIR}"
    echo "Custom CA imported into Firefox profile."
else
    echo "Warning: certutil not found. Cannot import CA into Firefox."
fi
