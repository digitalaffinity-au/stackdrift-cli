#!/usr/bin/env bash
set -euo pipefail

# Installs the StackDrift CLI on Linux and macOS. Downloads the release binary
# and places it in a directory that is already on your PATH, so no environment
# variable changes are needed. Run with:
#   curl -fsSL https://raw.githubusercontent.com/digitalaffinity-au/stackdrift-cli/main/scripts/install.sh | bash

REPO="digitalaffinity-au/stackdrift-cli"

echo "Installing the StackDrift CLI"

case "$(uname -s)" in
  Linux) OS="linux" ;;
  Darwin) OS="darwin" ;;
  *) echo "Unsupported operating system: $(uname -s)"; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $(uname -m)"; exit 1 ;;
esac

BINARY="stackdrift-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${BINARY}"

TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

echo "Downloading ${URL}"
curl -fsSL "$URL" -o "$TMP"
chmod +x "$TMP"

on_path() {
  case ":$PATH:" in
    *":$1:"*) return 0 ;;
    *) return 1 ;;
  esac
}

TARGET=""
USE_SUDO=""

if [ -n "${STACKDRIFT_INSTALL_DIR:-}" ]; then
  mkdir -p "$STACKDRIFT_INSTALL_DIR"
  TARGET="$STACKDRIFT_INSTALL_DIR/stackdrift"
else
  for dir in "$HOME/.local/bin" "$HOME/bin" "/opt/homebrew/bin" "/usr/local/bin"; do
    if on_path "$dir"; then
      mkdir -p "$dir" 2>/dev/null || true
      if [ -w "$dir" ]; then TARGET="$dir/stackdrift"; break; fi
    fi
  done

  if [ -z "$TARGET" ]; then
    saved_ifs="$IFS"
    set -f
    IFS=":"
    set -- $PATH
    set +f
    IFS="$saved_ifs"
    for dir in "$@"; do
      [ -n "$dir" ] || continue
      if [ -d "$dir" ] && [ -w "$dir" ]; then TARGET="$dir/stackdrift"; break; fi
    done
  fi

  if [ -z "$TARGET" ] && on_path "/usr/local/bin" && command -v sudo >/dev/null 2>&1; then
    TARGET="/usr/local/bin/stackdrift"
    USE_SUDO="yes"
  fi

  if [ -z "$TARGET" ]; then
    mkdir -p "$HOME/.local/bin"
    TARGET="$HOME/.local/bin/stackdrift"
  fi
fi

if [ -n "$USE_SUDO" ]; then
  echo "Installing to $TARGET (needs sudo)"
  sudo install -m 0755 "$TMP" "$TARGET"
else
  mkdir -p "$(dirname "$TARGET")"
  install -m 0755 "$TMP" "$TARGET"
fi

echo "Installed to ${TARGET}"

INSTALL_DIR="$(dirname "$TARGET")"
if on_path "$INSTALL_DIR"; then
  echo "That directory is already on your PATH."
else
  echo
  echo "That directory is not on your PATH. Add it, then open a new terminal:"
  echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
fi

echo
echo "Next: run 'stackdrift login' then 'stackdrift scan' in a project directory."
