#!/bin/sh
set -eu

CA_CERT="/certs/dev-root-ca.crt"

if [ ! -f "$CA_CERT" ]; then
  echo "No custom CA certificate found at $CA_CERT, skipping."
  exit 0
fi

echo "Installing custom CA into system trust store..."
mkdir -p /usr/local/share/ca-certificates
cp "$CA_CERT" /usr/local/share/ca-certificates/dev-root-ca.crt
update-ca-certificates >/dev/null 2>&1 || update-ca-certificates
