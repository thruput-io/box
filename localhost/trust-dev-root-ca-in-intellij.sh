#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "This script supports macOS only."
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_ROOT_DIR="${BOX_ROOT:-$(cd "${SCRIPT_DIR}/.." && pwd)}"

CA_CERT="${COMPOSE_ROOT_DIR}/certs/dev-root-ca.crt"
ALIAS_NAME="box-dev-root-ca"
STOREPASS="${JBR_CACERTS_STOREPASS:-changeit}"

if [[ ! -f "${CA_CERT}" ]]; then
  echo "Missing certificate: ${CA_CERT}"
  echo "Generate certificates first (for example: make -C ${COMPOSE_ROOT_DIR} generate-certs)."
  exit 1
fi

function find_cacerts_candidates() {
  local candidates=()

  # Common direct install locations
  if [[ -f "/Applications/IntelliJ IDEA.app/Contents/jbr/Contents/Home/lib/security/cacerts" ]]; then
    candidates+=("/Applications/IntelliJ IDEA.app/Contents/jbr/Contents/Home/lib/security/cacerts")
  fi
  if [[ -f "$HOME/Applications/IntelliJ IDEA.app/Contents/jbr/Contents/Home/lib/security/cacerts" ]]; then
    candidates+=("$HOME/Applications/IntelliJ IDEA.app/Contents/jbr/Contents/Home/lib/security/cacerts")
  fi

  # JetBrains Toolbox installs (most common)
  if command -v mdfind >/dev/null 2>&1; then
    while IFS= read -r p; do
      [[ -n "${p}" ]] && candidates+=("${p}")
    done < <(mdfind "kMDItemFSName == 'cacerts' && kMDItemPath CONTAINS 'IntelliJ IDEA.app/Contents/jbr/Contents/Home/lib/security'" 2>/dev/null | head -n 50)
  fi

  # De-dup
  local out=()
  local seen=""
  for p in "${candidates[@]:-}"; do
    if [[ -f "${p}" ]] && [[ "${seen}" != *"|${p}|"* ]]; then
      out+=("${p}")
      seen+="|${p}|"
    fi
  done

  printf '%s\n' "${out[@]:-}"
}

function pick_cacerts() {
  local picked=""
  while IFS= read -r p; do
    [[ -n "${p}" ]] || continue
    picked="${p}"
    break
  done < <(find_cacerts_candidates)

  if [[ -z "${picked}" ]]; then
    echo "Unable to locate IntelliJ JBR cacerts automatically."
    echo "If you know the path, set JBR_CACERTS_PATH and re-run."
    echo "Example:"
    echo "  JBR_CACERTS_PATH=\"/Applications/IntelliJ IDEA.app/Contents/jbr/Contents/Home/lib/security/cacerts\" $0"
    exit 1
  fi

  echo "${picked}"
}

CACHE_PATH="${JBR_CACERTS_PATH:-}"
if [[ -z "${CACHE_PATH}" ]]; then
  CACHE_PATH="$(pick_cacerts)"
fi

JBR_HOME="$(cd "$(dirname "${CACHE_PATH}")/../.." && pwd)"
KEYTOOL_DEFAULT="${JBR_HOME}/bin/keytool"

KEYTOOL_BIN=""
if command -v keytool >/dev/null 2>&1; then
  KEYTOOL_BIN="$(command -v keytool)"
elif [[ -x "${KEYTOOL_DEFAULT}" ]]; then
  KEYTOOL_BIN="${KEYTOOL_DEFAULT}"
else
  echo "Unable to find 'keytool' in PATH or under IntelliJ JBR."
  echo "Tried: ${KEYTOOL_DEFAULT}"
  exit 1
fi

echo "Using cacerts: ${CACHE_PATH}"
echo "Using keytool: ${KEYTOOL_BIN}"

function has_alias() {
  "${KEYTOOL_BIN}" -list -keystore "${CACHE_PATH}" -storepass "${STOREPASS}" -alias "${ALIAS_NAME}" >/dev/null 2>&1
}

if has_alias; then
  echo "CA already trusted in IntelliJ JBR (alias '${ALIAS_NAME}')."
  exit 0
fi

IMPORT_CMD=("${KEYTOOL_BIN}" -importcert -noprompt -trustcacerts -alias "${ALIAS_NAME}" -file "${CA_CERT}" -keystore "${CACHE_PATH}" -storepass "${STOREPASS}")

if [[ -w "${CACHE_PATH}" ]]; then
  "${IMPORT_CMD[@]}"
else
  echo "cacerts is not writable; attempting import with sudo..."
  sudo "${IMPORT_CMD[@]}"
fi

echo "Imported ${CA_CERT} into IntelliJ JBR truststore. Restart IntelliJ IDEA."
