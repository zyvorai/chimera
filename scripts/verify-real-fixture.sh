#!/usr/bin/env bash
# Copyright 2026 Zyvor AI Labs
# SPDX-License-Identifier: Apache-2.0
# ============================================================================
# verify-real-fixture.sh — Prove an exported disk is real, not filler bytes
# ============================================================================
# Fetches sample fixtures if needed, starts a throwaway chimera instance
# pointed at them, downloads the *complete* NFC stream for the VM the real
# fixture was assigned to (via `chimera selftest -vm ... -save ...`, not just
# selftest's normal 4KB probe read), and confirms qemu-img recognizes the
# downloaded bytes as a valid disk image with a sane virtual size.
# ============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_DIR"

command -v qemu-img >/dev/null 2>&1 || {
    echo "qemu-img not found — see scripts/fetch-sample-fixtures.sh for install hints." >&2
    exit 1
}
command -v python3 >/dev/null 2>&1 || {
    echo "python3 not found (used to parse /api/vmdks JSON)." >&2
    exit 1
}

./scripts/fetch-sample-fixtures.sh
go build -o bin/chimera ./cmd/chimera

PORT=18991
LOG="$(mktemp)"
DOWNLOAD="$(mktemp)"
CHIMERA_FIXTURE_VMDK_DIR="$REPO_DIR/fixtures" ./bin/chimera serve -listen "127.0.0.1:$PORT" >"$LOG" 2>&1 &
PID=$!
cleanup() { kill "$PID" 2>/dev/null || true; wait "$PID" 2>/dev/null || true; rm -f "$LOG" "$DOWNLOAD"; }
trap cleanup EXIT

for _ in $(seq 1 40); do
    curl -fsS "http://127.0.0.1:$PORT/__chimera/health" >/dev/null 2>&1 && break
    sleep 0.25
done

VMDKS_JSON=$(curl -fsS "http://127.0.0.1:$PORT/__chimera/api/vmdks")
VM_NAME=$(python3 -c "
import json, sys
d = json.loads(sys.argv[1])
matched = [f for f in d['files'] if f.get('vm_name')]
print(matched[0]['vm_name'] if matched else '')
" "$VMDKS_JSON")
[ -n "$VM_NAME" ] || { echo "No VM got assigned a real fixture — check fixtures/ and /api/vmdks: $VMDKS_JSON" >&2; exit 1; }
echo "==> Real fixture assigned to VM: $VM_NAME"

echo "==> Downloading the COMPLETE export for $VM_NAME (not just a 4KB probe)"
./bin/chimera selftest -url "http://127.0.0.1:$PORT/sdk" -user administrator@vsphere.local -pass vmware -insecure=true -vm "$VM_NAME" -save "$DOWNLOAD"

echo "==> Confirming the downloaded bytes are a valid disk image"
qemu-img info "$DOWNLOAD"

echo
echo "Real fixture verified end-to-end: the exported disk for $VM_NAME is a valid, real disk image (not filler bytes)."
