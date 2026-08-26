#!/usr/bin/env bash
set -euo pipefail
TRANSIVA_DIR="${TRANSIVA_DIR:-../transiva}"
LAB_URL="${CHIMERA_URL:-http://127.0.0.1:8989/sdk}"
USER="${CHIMERA_USERNAME:-administrator@vsphere.local}"
PASS="${CHIMERA_PASSWORD:-vmware}"
OUT="${OUT:-/tmp/transiva-chimera}"
CFG="${CFG:-/tmp/transiva-chimera.yaml}"

curl -fsS "${LAB_URL%/sdk}/__chimera/health" >/dev/null
CHIMERA_URL="$LAB_URL" CHIMERA_USERNAME="$USER" CHIMERA_PASSWORD="$PASS" ./scripts/make-transiva-config.sh > "$CFG"
mkdir -p "$OUT"

echo "Lab endpoint: $LAB_URL"
echo "Config:       $CFG"
echo "Output:       $OUT"
echo
echo "1) Start with the Chimera self-test:"
echo "   ./bin/chimera selftest -url '$LAB_URL' -user '$USER' -pass '$PASS'"
echo
echo "2) Then run Transiva interactive export:"
echo "   $TRANSIVA_DIR/bin/hyperexport --config '$CFG'"
echo
echo "Choose VM: DC0_C0_RP0_VM0 (or list the generated VMs in the TUI)."
