#!/usr/bin/env bash
# ============================================================================
# deploy-remote.sh — Deploy Chimera to a remote host as a systemd service
# ============================================================================
# Chimera builds as a single static binary (CGO_ENABLED=0, see Dockerfile), so
# deployment needs no remote toolchain (no Go/pip/npm install on the target):
#   1. Detect the remote OS/arch over SSH
#   2. Cross-compile ./cmd/chimera for that target, locally
#   3. Copy the binary + an env file + a systemd unit to the remote host
#   4. Enable/start chimera.service, open the firewall port
#   5. Verify with `chimera selftest` (login, inventory, ExportVm, NFC lease)
#   6. Point scripts/transiva-smoke.sh at the deployed instance
#
# Usage:
#   ./scripts/deploy-remote.sh <host> [user] [password] [options]
#   ./scripts/deploy-remote.sh 80.79.5.173 sus                 # SSH key auth
#   ./scripts/deploy-remote.sh 80.79.5.173 sus mypassword      # SSH password auth
#   ./scripts/deploy-remote.sh 80.79.5.173 sus --uninstall
#   ./scripts/deploy-remote.sh 80.79.5.173 sus --dry-run
#
# Options:
#   --uninstall   Stop chimera.service and remove it + the binary from the host
#   --dry-run     Print what would happen; make no changes
#   --skip-smoke  Skip the transiva-smoke.sh step at the end
#   --verbose     Show full remote command output
#
# Environment variables:
#   DEPLOY_HOST, DEPLOY_USER, DEPLOY_PASS   — same as the positional args
#   CHIMERA_PORT                             — listen/health-check port (default 8989)
#   CHIMERA_USERNAME, CHIMERA_PASSWORD, CHIMERA_ADMIN_TOKEN, CHIMERA_VMS_PER_POOL,
#   CHIMERA_FIXTURE_SIZE_MB                  — seed values for the remote env file
#                                               (see internal/config/config.go for
#                                               the full CHIMERA_* list; only written
#                                               once, at /etc/chimera/chimera.env —
#                                               re-running deploy never overwrites it)
# ============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=lib/deploy-common.sh
source "$SCRIPT_DIR/lib/deploy-common.sh"

info()  { chimera_info "$@"; }
warn()  { chimera_warn "$@"; }
error() { chimera_error "$@"; }
step()  { LAST_ACTION="$*"; deploy_ui_step_start "$*"; }

REMOTE_BIN=/usr/local/bin/chimera
REMOTE_ENV=/etc/chimera/chimera.env
REMOTE_UNIT=/etc/systemd/system/chimera.service
CHIMERA_PORT="${CHIMERA_PORT:-8989}"

# ── Parse args ──
UNINSTALL_MODE=false
DRY_RUN=false
SKIP_SMOKE=false
VERBOSE=false
POSITIONAL=()
for arg in "$@"; do
    case "$arg" in
        --uninstall)  UNINSTALL_MODE=true ;;
        --dry-run)    DRY_RUN=true ;;
        --skip-smoke) SKIP_SMOKE=true ;;
        --verbose)    VERBOSE=true ;;
        --help|-h)
            sed -n '2,32p' "$0" | sed 's/^# \{0,1\}//'
            exit 0
            ;;
        *) POSITIONAL+=("$arg") ;;
    esac
done

HOST="${POSITIONAL[0]:-${DEPLOY_HOST:-}}"
USER="${POSITIONAL[1]:-${DEPLOY_USER:-root}}"
PASS="${POSITIONAL[2]:-${DEPLOY_PASS:-}}"

chimera_parse_target HOST USER
if [ -z "$HOST" ] && chimera_load_deploy_last "$REPO_DIR"; then
    info "Using .deploy-last → ${USER}@${HOST}"
fi
[ -z "$HOST" ] && error "Usage: $0 <host> [user] [password] [options]  (see --help)"

[ -f "$REPO_DIR/go.mod" ] || error "Not in the chimera repo: $REPO_DIR"
chimera_build_metadata "$REPO_DIR"

SUDO=""
[ "$USER" != "root" ] && SUDO="sudo"

DEPLOY_SSH_OPTS=(
    -o StrictHostKeyChecking=no
    -o ConnectTimeout=15
    -o ServerAliveInterval=15
    -o ServerAliveCountMax=8
)
DEPLOY_SSH_TTY_OPTS=()
[ "$USER" != "root" ] && DEPLOY_SSH_TTY_OPTS=(-tt)

if [ -n "$PASS" ] && ! command -v sshpass &>/dev/null; then
    error "sshpass required for password auth (brew install sshpass / dnf install sshpass)"
fi

_ssh() {
    local -a ssh_args=("${DEPLOY_SSH_OPTS[@]}" "${DEPLOY_SSH_TTY_OPTS[@]}")
    if [ -n "$PASS" ]; then
        SSHPASS="$PASS" sshpass -e ssh "${ssh_args[@]}" "${USER}@${HOST}" "$@"
    else
        ssh "${ssh_args[@]}" "${USER}@${HOST}" "$@"
    fi
}

# Short probes — no TTY (avoids spurious "Connection closed" between quick commands).
_ssh_batch() {
    local -a ssh_args=("${DEPLOY_SSH_OPTS[@]}")
    if [ -n "$PASS" ]; then
        SSHPASS="$PASS" sshpass -e ssh "${ssh_args[@]}" "${USER}@${HOST}" "$@"
    else
        ssh "${ssh_args[@]}" "${USER}@${HOST}" "$@"
    fi
}

_scp() {
    local -a scp_args=("${DEPLOY_SSH_OPTS[@]}")
    if [ -n "$PASS" ]; then
        SSHPASS="$PASS" sshpass -e scp "${scp_args[@]}" "$@"
    else
        scp "${scp_args[@]}" "$@"
    fi
}

if $DRY_RUN; then
    deploy_ui_banner "${DEPLOY_UI_ICON_MAGIC} Dry run" "no changes will be made"
    deploy_ui_kv "🎯" "Target" "${USER}@${HOST}"
    deploy_ui_kv "📦" "Binary" "$REMOTE_BIN"
    deploy_ui_kv "📄" "Env file" "$REMOTE_ENV (written only if missing)"
    deploy_ui_kv "⚙️" "Unit" "$REMOTE_UNIT"
    echo ""
    deploy_ui_note "Would: detect arch → cross-compile locally → scp binary+unit → enable/start → selftest → transiva-smoke.sh"
    echo ""
    exit 0
fi

deploy_ui_banner "Remote Deploy" "${CHIMERA_GIT_VERSION} (${CHIMERA_GIT_COMMIT}) → ${USER}@${HOST}"
deploy_ui_kv "🎯" "Target" "${USER}@${HOST}"
deploy_ui_kv "🔐" "Auth" "$([ -n "$PASS" ] && echo 'password' || echo 'SSH key')"
deploy_ui_kv "🌐" "Port" "$CHIMERA_PORT"
echo ""

# ── Uninstall mode ──
if $UNINSTALL_MODE; then
    deploy_ui_uninstall_banner
    step "Uninstalling chimera from ${HOST}"
    _ssh "
        $SUDO systemctl disable --now chimera.service 2>/dev/null || true
        $SUDO rm -f $REMOTE_BIN $REMOTE_UNIT
        $SUDO rm -rf /etc/chimera
        $SUDO systemctl daemon-reload 2>/dev/null || true
    "
    info "chimera removed from ${HOST}"
    exit 0
fi

# ── Step 1: detect remote OS/arch ──
step "Detecting remote architecture"
REMOTE_ARCH_RAW=$(_ssh_batch "uname -m" | tr -d '\r')
case "$REMOTE_ARCH_RAW" in
    x86_64)         GOARCH=amd64 ;;
    aarch64|arm64)  GOARCH=arm64 ;;
    *) error "Unsupported remote architecture: $REMOTE_ARCH_RAW" ;;
esac
info "Remote: linux/${GOARCH}"

# ── Step 2: cross-compile locally (static binary, no CGO — see Dockerfile) ──
step "Cross-compiling chimera for linux/${GOARCH}"
BUILD_DIR="$(mktemp -d)"
trap 'rm -rf "$BUILD_DIR"' EXIT
(
    cd "$REPO_DIR"
    CGO_ENABLED=0 GOOS=linux GOARCH="$GOARCH" \
        go build -trimpath -ldflags="-s -w" -o "$BUILD_DIR/chimera" ./cmd/chimera
)
info "Built $BUILD_DIR/chimera"

# ── Step 3: install binary + env file + systemd unit ──
step "Installing chimera on ${HOST}"
_ssh "$SUDO mkdir -p /etc/chimera"
_scp "$BUILD_DIR/chimera" "${USER}@${HOST}:/tmp/chimera.new"
_ssh "$SUDO install -m 755 /tmp/chimera.new $REMOTE_BIN && rm -f /tmp/chimera.new"

cat > "$BUILD_DIR/chimera.env" <<ENVEOF
CHIMERA_LISTEN=0.0.0.0:${CHIMERA_PORT}
CHIMERA_PUBLIC_HOST=${HOST}:${CHIMERA_PORT}
CHIMERA_USERNAME=${CHIMERA_USERNAME:-administrator@vsphere.local}
CHIMERA_PASSWORD=${CHIMERA_PASSWORD:-vmware}
CHIMERA_ADMIN_TOKEN=${CHIMERA_ADMIN_TOKEN:-chimera-admin}
CHIMERA_VMS_PER_POOL=${CHIMERA_VMS_PER_POOL:-4}
CHIMERA_FIXTURE_SIZE_MB=${CHIMERA_FIXTURE_SIZE_MB:-16}
ENVEOF
_scp "$BUILD_DIR/chimera.env" "${USER}@${HOST}:/tmp/chimera.env.new"
# Never clobber an operator's existing env file on redeploy — only the binary/unit update.
_ssh "
    if [ ! -f $REMOTE_ENV ]; then
        $SUDO install -m 640 /tmp/chimera.env.new $REMOTE_ENV
    fi
    rm -f /tmp/chimera.env.new
"

cat > "$BUILD_DIR/chimera.service" <<'UNITEOF'
[Unit]
Description=Chimera infrastructure endpoint simulator
Documentation=https://github.com/zyvorai/chimera
After=network.target

[Service]
Type=simple
EnvironmentFile=-/etc/chimera/chimera.env
ExecStart=/usr/local/bin/chimera serve
Restart=on-failure
RestartSec=5s
DynamicUser=yes
StateDirectory=chimera
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
StandardOutput=journal
StandardError=journal
SyslogIdentifier=chimera

[Install]
WantedBy=multi-user.target
UNITEOF
_scp "$BUILD_DIR/chimera.service" "${USER}@${HOST}:/tmp/chimera.service.new"
_ssh "$SUDO install -m 644 /tmp/chimera.service.new $REMOTE_UNIT && rm -f /tmp/chimera.service.new"
info "Binary + config + systemd unit installed"

# ── Step 4: enable/start service, open firewall ──
step "Starting chimera.service"
_ssh "
    $SUDO systemctl daemon-reload
    $SUDO systemctl enable --now chimera.service
    $SUDO systemctl restart chimera.service
    if command -v firewall-cmd &>/dev/null; then
        $SUDO firewall-cmd --permanent --add-port=${CHIMERA_PORT}/tcp 2>/dev/null || true
        $SUDO firewall-cmd --reload 2>/dev/null || true
    elif command -v ufw &>/dev/null; then
        $SUDO ufw allow ${CHIMERA_PORT}/tcp 2>/dev/null || true
    fi
    sleep 1
    if $SUDO systemctl is-active chimera.service &>/dev/null; then
        echo 'chimera.service: running'
    else
        echo 'chimera.service: FAILED TO START'
        $SUDO journalctl -u chimera.service --no-pager -n 20
        exit 1
    fi
"
info "chimera.service active"

# ── Step 5: verify with the real vSphere contract (login, inventory, export, NFC) ──
step "Verifying deployment"
# CHIMERA_TLS may already be set in the remote chimera.env from a prior
# manual change (deploy never touches that file) — probe rather than assume,
# so verification doesn't false-fail against a TLS-only listener.
LAB_SCHEME="http"
if _ssh_batch "curl -sk -o /dev/null -w '%{http_code}' https://127.0.0.1:${CHIMERA_PORT}/__chimera/health" 2>/dev/null | grep -q '^200$'; then
    LAB_SCHEME="https"
fi
LAB_URL="${LAB_SCHEME}://${HOST}:${CHIMERA_PORT}/sdk"
DEPLOY_UI_SCHEME="$LAB_SCHEME" # so the final success banner prints the right URL scheme
_ssh "curl -fsSk ${LAB_SCHEME}://127.0.0.1:${CHIMERA_PORT}/__chimera/health >/dev/null" \
    && info "Health check OK (${LAB_SCHEME}://127.0.0.1:${CHIMERA_PORT}/__chimera/health, on-host)"

LOCAL_BIN="$REPO_DIR/bin/chimera"
if [ ! -x "$LOCAL_BIN" ]; then
    ( cd "$REPO_DIR" && go build -o bin/chimera ./cmd/chimera )
fi
"$LOCAL_BIN" selftest -url "$LAB_URL" \
    -user "${CHIMERA_USERNAME:-administrator@vsphere.local}" \
    -pass "${CHIMERA_PASSWORD:-vmware}" \
    -insecure=true \
    && info "selftest OK — login, inventory, ExportVm and NFC lease all verified against ${LAB_URL}"

chimera_save_deploy_last "$REPO_DIR" "$HOST" "$USER" "full"

deploy_ui_highlight "📋 Final checklist"
deploy_ui_checklist "service" "$(_ssh_batch "$SUDO systemctl is-active chimera.service" | tr -d '\r')"
deploy_ui_checklist "health"  "$(_ssh_batch "curl -sk -o /dev/null -w '%{http_code}' ${LAB_SCHEME}://127.0.0.1:${CHIMERA_PORT}/__chimera/health" | tr -d '\r')"

chimera_print_success "$HOST" 0

# ── Step 6: wire up the transiva integration test ──
if $SKIP_SMOKE; then
    info "Skipped transiva-smoke.sh (--skip-smoke)"
else
    step "Pointing scripts/transiva-smoke.sh at ${HOST}"
    ( cd "$REPO_DIR" && CHIMERA_URL="$LAB_URL" \
        CHIMERA_USERNAME="${CHIMERA_USERNAME:-administrator@vsphere.local}" \
        CHIMERA_PASSWORD="${CHIMERA_PASSWORD:-vmware}" \
        ./scripts/transiva-smoke.sh )
fi
