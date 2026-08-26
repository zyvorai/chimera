#!/bin/sh
set -e
systemctl daemon-reload >/dev/null 2>&1 || true
echo "chimera installed. Start it with: systemctl enable --now chimera"
