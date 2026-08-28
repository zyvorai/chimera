#!/usr/bin/env bash
# Copyright 2026 Zyvor AI Labs
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
BASE="${CHIMERA_BASE:-http://127.0.0.1:8989}"
TOKEN="${CHIMERA_ADMIN_TOKEN:-chimera-admin}"
NAME="${1:-clean}"
curl -fsS -X POST -H "Authorization: Bearer $TOKEN" "$BASE/__chimera/scenario/$NAME"
echo
