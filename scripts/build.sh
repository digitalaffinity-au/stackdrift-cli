#!/usr/bin/env bash
set -euo pipefail

# Builds release binaries into dist/ for linux, windows, and macOS on amd64 and
# arm64. The server URL is baked in at compile time. Pass a version as the first
# argument.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-dev}"
SERVER_URL="${STACKDRIFT_BUILD_URL:-https://stackdrift.net}"
MODULE="github.com/digitalaffinity-au/stackdrift-cli"

LDFLAGS="-s -w -X main.version=${VERSION} -X ${MODULE}/internal/config.DefaultBaseURL=${SERVER_URL}"

mkdir -p "$ROOT/dist"

echo "Building stackdrift ${VERSION} for server ${SERVER_URL}"

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
