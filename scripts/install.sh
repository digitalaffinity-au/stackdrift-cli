#!/usr/bin/env bash
set -euo pipefail

# Installs the StackDrift CLI on Linux. Downloads the release binary and puts it
# on your PATH. Run with:
#   curl -fsSL https://raw.githubusercontent.com/digitalaffinity-au/stackdrift-cli/main/scripts/install.sh | bash

REPO="digitalaffinity-au/stackdrift-cli"
INSTALL_DIR="${STACKDRIFT_INSTALL_DIR:-$HOME/.local/bin}"
TARGET="$INSTALL_DIR/stackdrift"

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

mkdir -p "$INSTALL_DIR"

URL="https://github.com/${REPO}/releases/latest/download/${BINARY}"
echo "Downloading ${URL}"
curl -fsSL "$URL" -o "$TARGET"
chmod +x "$TARGET"

echo "Installed to ${TARGET}"

if ! command -v stackdrift >/dev/null 2>&1; then
  echo
  echo "Add this directory to your PATH, then open a new terminal:"
  echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
fi

echo
echo "Next: run 'stackdrift login' then 'stackdrift scan' in a project directory."
