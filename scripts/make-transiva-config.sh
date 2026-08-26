#!/usr/bin/env sh
set -eu
URL="${CHIMERA_URL:-http://127.0.0.1:8989/sdk}"
USER="${CHIMERA_USERNAME:-administrator@vsphere.local}"
PASS="${CHIMERA_PASSWORD:-vmware}"
cat <<YAML
vcenter_url: "$URL"
username: "$USER"
password: "$PASS"
insecure: true
timeout: 10m
download_workers: 2
retry_attempts: 4
retry_delay: 1s
log_level: debug
progress_style: bar
show_eta: true
refresh_rate: 100ms
YAML
