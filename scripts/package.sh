#!/usr/bin/env bash
# ============================================================================
# package.sh — Build .deb and .rpm packages for chimera
# ============================================================================
# Chimera is a single static binary (CGO_ENABLED=0), so packaging is just:
#   1. Cross-compile ./cmd/chimera for each target arch
#   2. Run nfpm against packaging/nfpm.yaml to produce a .deb and .rpm
#
# Usage:
#   ./scripts/package.sh                  # both arches, both formats
#   ./scripts/package.sh amd64             # one arch, both formats
#   ./scripts/package.sh amd64 deb         # one arch, one format
#
# Requires nfpm (https://nfpm.goreleaser.com):
#   go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
# ============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_DIR"

command -v nfpm >/dev/null 2>&1 || {
    echo "nfpm not found. Install with: go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest" >&2
    exit 1
}

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo dev)}"
ARCHES="${1:-amd64 arm64}"
FORMATS="${2:-deb rpm}"

mkdir -p dist dist/build
for ARCH in $ARCHES; do
    echo "==> Building linux/${ARCH} binary (chimera ${VERSION})"
    CGO_ENABLED=0 GOOS=linux GOARCH="$ARCH" \
        go build -trimpath -ldflags="-s -w" -o dist/build/chimera ./cmd/chimera

    for FORMAT in $FORMATS; do
        echo "==> Packaging ${FORMAT}/${ARCH}"
        VERSION="$VERSION" ARCH="$ARCH" nfpm package \
            --config packaging/nfpm.yaml \
            --packager "$FORMAT" \
            --target "dist/"
    done
done

echo
echo "==> Built packages:"
ls -la dist/
