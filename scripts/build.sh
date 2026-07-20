#!/usr/bin/env bash
set -euo pipefail

# Builds release binaries into dist/ for linux, windows, and macOS on amd64 and
# arm64. All binaries point at https://stackdrift.net; the only way to target a
# different server is the STACKDRIFT_URL environment variable at runtime. Pass a
# version as the first argument.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  if [ -f "$ROOT/VERSION" ]; then
    VERSION="$(tr -d ' \t\n\r' < "$ROOT/VERSION")"
  else
    VERSION="dev"
  fi
fi

LDFLAGS="-s -w -X main.version=${VERSION}"

mkdir -p "$ROOT/dist"

echo "Building stackdrift ${VERSION} (server: https://stackdrift.net, override with STACKDRIFT_URL)"

build() {
  local goos="$1" goarch="$2" ext="$3"
  local out="$ROOT/dist/stackdrift-${goos}-${goarch}${ext}"
  GOOS="$goos" GOARCH="$goarch" go build -ldflags "$LDFLAGS" -o "$out" "$ROOT/cmd/stackdrift"
  echo "  dist/stackdrift-${goos}-${goarch}${ext}"
}

build linux amd64 ""
build linux arm64 ""
build windows amd64 ".exe"
build windows arm64 ".exe"
build darwin amd64 ""
build darwin arm64 ""

echo "Done."
