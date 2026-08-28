#!/bin/sh
# Copyright 2026 Zyvor AI Labs
# SPDX-License-Identifier: Apache-2.0
set -e
systemctl daemon-reload >/dev/null 2>&1 || true
echo "chimera installed. Start it with: systemctl enable --now chimera"
