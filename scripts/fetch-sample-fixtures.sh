#!/usr/bin/env bash
# ============================================================================
# fetch-sample-fixtures.sh — Fetch real, small, bootable Linux disk images
# ============================================================================
# By default Chimera's exported "disk" is a deterministic filler byte stream
# (see internal/fixture/store.go) — fine for testing the transfer/retry path,
# but there's nothing real to boot, convert, or inspect afterward. This script
# downloads Alpine Linux's official "tiny" cloud image (small, legitimate,
# vendor-published and checksummed) and converts it to .vmdk, so
# CHIMERA_FIXTURE_VMDK_DIR can point at real disks out of the box instead of
# requiring the operator to supply their own.
#
# Usage:
#   ./scripts/fetch-sample-fixtures.sh [output-dir]   # default: ./fixtures
# ============================================================================

set -euo pipefail

OUT_DIR="${1:-fixtures}"
INDEX_URL="https://dl-cdn.alpinelinux.org/alpine/latest-stable/releases/cloud/"

command -v qemu-img >/dev/null 2>&1 || {
    echo "qemu-img not found. Install it first:" >&2
    echo "  macOS:  brew install qemu" >&2
    echo "  Fedora/RHEL/Alma: sudo dnf install qemu-img" >&2
    echo "  Debian/Ubuntu: sudo apt install qemu-utils" >&2
    exit 1
}

mkdir -p "$OUT_DIR"
VMDK_OUT="$OUT_DIR/alpine-virt.vmdk"
if [ -f "$VMDK_OUT" ]; then
    echo "Already have $VMDK_OUT — skipping (delete it to re-fetch)."
    exit 0
fi

echo "==> Finding the current Alpine 'tiny' cloud image at $INDEX_URL"
INDEX_HTML=$(curl -fsS "$INDEX_URL")
QCOW2_NAME=$(printf '%s' "$INDEX_HTML" | grep -Eo 'href="generic_alpine-[0-9.]+-x86_64-bios-tiny-r[0-9]+\.qcow2"' | sed -E 's/^href="(.*)"$/\1/' | sort -u | tail -1)
[ -n "$QCOW2_NAME" ] || {
    echo "Could not find a generic_alpine-*-x86_64-bios-tiny-*.qcow2 image at $INDEX_URL — Alpine may have changed its layout." >&2
    echo "Browse $INDEX_URL manually and pass a direct URL, or supply your own VMDK via CHIMERA_FIXTURE_VMDK_DIR." >&2
    exit 1
}

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT
QCOW2_PATH="$TMP_DIR/$QCOW2_NAME"

echo "==> Downloading $QCOW2_NAME"
curl -fSL --progress-bar -o "$QCOW2_PATH" "$INDEX_URL$QCOW2_NAME"

echo "==> Verifying checksum"
EXPECTED_SHA512=$(curl -fsS "$INDEX_URL$QCOW2_NAME.sha512" | awk '{print $1}')
if command -v sha512sum >/dev/null 2>&1; then
    ACTUAL_SHA512=$(sha512sum "$QCOW2_PATH" | awk '{print $1}')
else
    ACTUAL_SHA512=$(shasum -a 512 "$QCOW2_PATH" | awk '{print $1}')
fi
[ "$EXPECTED_SHA512" = "$ACTUAL_SHA512" ] || {
    echo "Checksum mismatch for $QCOW2_NAME! expected=$EXPECTED_SHA512 actual=$ACTUAL_SHA512" >&2
    exit 1
}
echo "Checksum OK"

echo "==> Converting to VMDK"
qemu-img convert -O vmdk "$QCOW2_PATH" "$VMDK_OUT"

echo
echo "==> Done: $VMDK_OUT"
qemu-img info "$VMDK_OUT"
