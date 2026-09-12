#!/usr/bin/env bash
# Reproducible Windows 7 build + packaging for the reduced Reasonix CLI.
#
# Windows 7 support is a REDUCED build: it is compiled with the Go 1.20
# toolchain (the last Go whose binaries run on Windows 7, golang/go#64622)
# using -tags win7 and the pinned go-1.20 module graph in go.win7.mod. The
# interactive TUI, MCP, and the pi model catalog are excluded; see
# dist/README-WIN7.txt for the full works/excluded list.
#
# Produces under dist/:
#   - reasonix-win7-amd64.exe   (64-bit Windows PE)
#   - reasonix-win7-amd64.zip   (exe + README-WIN7.txt, the portable deliverable)
#   - reasonix-win7-setup.exe   (only when makensis is available)
#
# 32-bit (GOARCH=386) is not produced: the pinned modernc.org/sqlite backend
# ships no 386 architecture support, so a 386 build does not link.
#
# Invoke via `make win7` (which validates the toolchain) or directly with
# WIN7_GOROOT pointing at a go1.20.x install.
set -euo pipefail

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

WIN7_GOROOT="${WIN7_GOROOT:-/projects/sandbox/toolchains/go1.20.14}"
GO="$WIN7_GOROOT/bin/go"

if [ ! -x "$GO" ]; then
  echo "build-win7: no Go toolchain at WIN7_GOROOT=$WIN7_GOROOT" >&2
  echo "build-win7: set WIN7_GOROOT to a go1.20.x install (last Win7-capable Go, golang/go#64622)" >&2
  exit 1
fi

# Refuse anything but go1.20.x: newer toolchains produce binaries that crash on
# launch on Windows 7 (golang/go#64622).
govers="$(GOTOOLCHAIN=local GOROOT="$WIN7_GOROOT" "$GO" env GOVERSION)"
case "$govers" in
  go1.20.*) : ;;
  *) echo "build-win7: WIN7_GOROOT is $govers, need go1.20.x (Go 1.21.5+ crashes on Win7)" >&2; exit 1 ;;
esac

# Version metadata: reuse the Makefile's pattern. Callers (make win7) pass these
# in; fall back to computing them so the script also works standalone.
VERSION="${VERSION:-$(git describe --tags --always 2>/dev/null || echo dev)}"
GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse --short=12 HEAD 2>/dev/null || echo unknown)}"
BUILD_TIME_UTC="${BUILD_TIME_UTC:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"
LDFLAGS="-s -w -X main.version=$VERSION -X main.gitCommit=$GIT_COMMIT -X main.buildTimeUTC=$BUILD_TIME_UTC"

mkdir -p dist

echo "== building dist/reasonix-win7-amd64.exe with $govers =="
GOTOOLCHAIN=local GOROOT="$WIN7_GOROOT" PATH="$WIN7_GOROOT/bin:$PATH" \
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
  "$GO" build -modfile=go.win7.mod -tags win7 -ldflags "$LDFLAGS" \
  -o dist/reasonix-win7-amd64.exe ./cmd/reasonix

echo "== packaging dist/reasonix-win7-amd64.zip =="
cp scripts/README-WIN7.txt dist/README-WIN7.txt
( cd dist && rm -f reasonix-win7-amd64.zip \
  && zip -j reasonix-win7-amd64.zip reasonix-win7-amd64.exe README-WIN7.txt )

# Lean CLI installer, only when makensis is available. This is deliberately NOT
# the heavy desktop NSIS pipeline (desktop/build/windows/installer): it just
# drops the single exe into Program Files and puts it on PATH.
if command -v makensis >/dev/null 2>&1; then
  echo "== building dist/reasonix-win7-setup.exe with makensis =="
  makensis -V2 -DVERSION="$VERSION" scripts/win7-installer.nsi
else
  echo "== makensis not found: skipping installer; the portable zip is the deliverable =="
fi

echo "== done =="
ls -la dist/
