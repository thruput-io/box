#!/usr/bin/env bash
set -euo pipefail

BOX_ROOT="${1:-}"

if [[ -z "${BOX_ROOT}" ]]; then
  echo "Usage: $(basename "$0") <compose-root-dir>"
  exit 1
fi

# Resolve to absolute path
BOX_ROOT="$(cd "${BOX_ROOT}" && pwd)"

BOX_SCRIPT_PATH="${HOME}/.local/bin/box"
BOX_SOURCE_SCRIPT="${BOX_ROOT}/localhost/box.sh"

if [[ ! -f "${BOX_SOURCE_SCRIPT}" ]]; then
  echo "Missing source script: ${BOX_SOURCE_SCRIPT}"
  exit 1
fi
mkdir -p "$(dirname "${BOX_SCRIPT_PATH}")"
cp "${BOX_SOURCE_SCRIPT}" "${BOX_SCRIPT_PATH}"
chmod +x "${BOX_SCRIPT_PATH}"

profile_file="${HOME}/.profile"
if [[ -n "${SHELL:-}" ]]; then
  shell_name="$(basename "${SHELL}")"
  if [[ -f "${HOME}/.${shell_name}rc" ]]; then
    profile_file="${HOME}/.${shell_name}rc"
  fi
fi

if ! grep -Fq 'export PATH="$HOME/.local/bin:$PATH"' "${profile_file}"; then
  printf '\nexport PATH="$HOME/.local/bin:$PATH"\n' >> "${profile_file}"
fi

if ! grep -Fq 'export BOX_ROOT=' "${profile_file}"; then
  printf 'export BOX_ROOT="%s"\n' "${BOX_ROOT}" >> "${profile_file}"
else
  awk -v box_root="${BOX_ROOT}" '
    BEGIN { replaced = 0 }
    /^export BOX_ROOT=/ {
      if (!replaced) {
        print "export BOX_ROOT=\"" box_root "\""
        replaced = 1
      }
      next
    }
    { print }
  ' "${profile_file}" > "${profile_file}.tmp"
  mv "${profile_file}.tmp" "${profile_file}"
fi

# --- Add shell completion for box command ---
if [[ -n "${SHELL:-}" ]]; then
  shell_name="$(basename "${SHELL}")"
  rc_file="${HOME}/.${shell_name}rc"
  if [[ -f "${rc_file}" ]]; then
    if [[ "${shell_name}" == "zsh" ]]; then
      if ! grep -Fq '_box_completion()' "${rc_file}"; then
        cat << 'EOF' >> "${rc_file}"

# box command Zsh completion
_box_completion() {
  local -a targets
  if [[ -n "$BOX_ROOT" && -f "$BOX_ROOT/Makefile" ]]; then
    targets=($(grep -E '^[a-zA-Z0-9_-]+:' "$BOX_ROOT/Makefile" | cut -d: -f1))
    _describe 'targets' targets
  fi
}
compdef _box_completion box
EOF
        echo "Added Zsh completion to ${rc_file}"
      fi
    elif [[ "${shell_name}" == "bash" ]]; then
      if ! grep -Fq '_box_completion()' "${rc_file}"; then
        cat << 'EOF' >> "${rc_file}"

# box command Bash completion
_box_completion() {
  local cur targets
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  if [[ -n "$BOX_ROOT" && -f "$BOX_ROOT/Makefile" ]]; then
    targets=$(grep -E '^[a-zA-Z0-9_-]+:' "$BOX_ROOT/Makefile" | cut -d: -f1)
    COMPREPLY=( $(compgen -W "${targets}" -- ${cur}) )
  fi
}
complete -F _box_completion box
EOF
        echo "Added Bash completion to ${rc_file}"
      fi
    fi
  fi
fi

echo "Installed box command at ${BOX_SCRIPT_PATH}"
echo "Run this once in your current shell: source ${profile_file}"
