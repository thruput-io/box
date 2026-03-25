#!/usr/bin/env bash
# Integration test for localhost/setup-dns-and-cert.sh and localhost/clean-dns-and-cert.sh
# Generates a temporary self-signed CA certificate, runs setup and clean,
# and verifies keychain + resolver state before and after each step.
#
# Requirements: macOS, openssl, security (Keychain), sudo access
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

SETUP_SCRIPT="${PROJECT_ROOT}/localhost/setup-dns-and-cert.sh"
CLEAN_SCRIPT="${PROJECT_ROOT}/localhost/clean-dns-and-cert.sh"

KEYCHAIN_PATH="${HOME}/Library/Keychains/infra-localhost.keychain-db"
LOGIN_KEYCHAIN="${HOME}/Library/Keychains/login.keychain-db"
RESOLVER_FILE="/etc/resolver/internal"

# --- helpers ---
PASS=0
FAIL=0
TESTS=0

assert() {
  local description="$1"
  local result="$2"  # 0 = pass
  TESTS=$((TESTS + 1))
  if [[ "${result}" -eq 0 ]]; then
    PASS=$((PASS + 1))
    echo "  ✓ ${description}"
  else
    FAIL=$((FAIL + 1))
    echo "  ✗ ${description}"
  fi
}

summary() {
  echo ""
  echo "Results: ${PASS}/${TESTS} passed, ${FAIL} failed"
  if [[ "${FAIL}" -gt 0 ]]; then
    exit 1
  fi
}

cert_in_keychain() {
  local keychain="$1"
  local sha1="$2"
  [[ -f "${keychain}" ]] && security find-certificate -a -Z "${keychain}" 2>/dev/null | tr -d ' ' | grep -Fq "${sha1}"
}

# --- generate a temporary test certificate ---
TMPDIR_TEST="$(mktemp -d)"
trap 'rm -rf "${TMPDIR_TEST}"' EXIT

TEST_CA_KEY="${TMPDIR_TEST}/test-ca-key.pem"
TEST_CA_CERT="${TMPDIR_TEST}/test-ca.pem"

openssl req -x509 -new -nodes -newkey rsa:2048 \
  -keyout "${TEST_CA_KEY}" -out "${TEST_CA_CERT}" \
  -days 1 -subj "/CN=localhost-test-ca" 2>/dev/null

TEST_CERT_SHA1="$(openssl x509 -in "${TEST_CA_CERT}" -noout -fingerprint -sha1 | awk -F= '{print $2}' | tr -d ':')"

# Copy test cert into the project certs dir (setup-dns-and-cert.sh reads from there)
CERT_DEST="${PROJECT_ROOT}/certs/dev-root-ca.crt"
CERT_BACKUP=""
if [[ -f "${CERT_DEST}" ]]; then
  CERT_BACKUP="${TMPDIR_TEST}/dev-root-ca.crt.bak"
  cp "${CERT_DEST}" "${CERT_BACKUP}"
fi
cp "${TEST_CA_CERT}" "${CERT_DEST}"

restore_cert() {
  if [[ -n "${CERT_BACKUP}" ]]; then
    cp "${CERT_BACKUP}" "${CERT_DEST}"
  else
    rm -f "${CERT_DEST}"
  fi
}
trap 'restore_cert; rm -rf "${TMPDIR_TEST}"' EXIT

# --- ensure clean starting state ---
echo "=== Ensuring clean starting state ==="
bash "${CLEAN_SCRIPT}" 2>/dev/null || true

echo ""
echo "=== Verify pre-setup state (everything should be absent) ==="

# Keychain should not exist
[[ ! -f "${KEYCHAIN_PATH}" ]]; assert "Custom keychain does not exist" $?

# Resolver file should not exist
[[ ! -f "${RESOLVER_FILE}" ]]; assert "Resolver file does not exist" $?

# Test cert should not be in login keychain
if [[ -f "${LOGIN_KEYCHAIN}" ]]; then
  ! cert_in_keychain "${LOGIN_KEYCHAIN}" "${TEST_CERT_SHA1}"; assert "Test cert not in login keychain" $?
else
  assert "Test cert not in login keychain (no login keychain)" 0
fi

# --- run setup ---
echo ""
echo "=== Running setup-dns-and-cert.sh ==="
bash "${SETUP_SCRIPT}"

echo ""
echo "=== Verify post-setup state ==="

# Custom keychain should exist
[[ -f "${KEYCHAIN_PATH}" ]]; assert "Custom keychain exists after setup" $?

# Test cert should be in custom keychain
cert_in_keychain "${KEYCHAIN_PATH}" "${TEST_CERT_SHA1}"; assert "Test cert in custom keychain" $?

# Resolver file should exist with correct content
[[ -f "${RESOLVER_FILE}" ]]; assert "Resolver file exists after setup" $?
grep -q "nameserver 127.0.0.1" "${RESOLVER_FILE}"; assert "Resolver has correct nameserver" $?
grep -q "port 5354" "${RESOLVER_FILE}"; assert "Resolver has correct port" $?

# Test cert should be in login keychain (Chrome trust)
if [[ -f "${LOGIN_KEYCHAIN}" ]]; then
  cert_in_keychain "${LOGIN_KEYCHAIN}" "${TEST_CERT_SHA1}"; assert "Test cert in login keychain (Chrome)" $?
else
  assert "Test cert in login keychain (Chrome) — skipped, no login keychain" 0
fi

# --- run clean ---
echo ""
echo "=== Running clean-dns-and-cert.sh ==="
bash "${CLEAN_SCRIPT}"

echo ""
echo "=== Verify post-clean state ==="

# Custom keychain should be removed
[[ ! -f "${KEYCHAIN_PATH}" ]]; assert "Custom keychain removed after clean" $?

# Resolver file should be removed
[[ ! -f "${RESOLVER_FILE}" ]]; assert "Resolver file removed after clean" $?

# Test cert should be removed from login keychain
if [[ -f "${LOGIN_KEYCHAIN}" ]]; then
  ! cert_in_keychain "${LOGIN_KEYCHAIN}" "${TEST_CERT_SHA1}"; assert "Test cert removed from login keychain" $?
else
  assert "Test cert removed from login keychain — skipped, no login keychain" 0
fi

# --- done ---
summary
