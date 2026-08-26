# SPDX-License-Identifier: Apache-2.0
# shellcheck shell=bash
# Chimera deploy library (self-contained under scripts/lib/).

_DEPLOY_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

DEPLOY_UI_PROJECT="chimera"
DEPLOY_UI_ICON="🧪"
DEPLOY_UI_ICON_UNINSTALL="🗑️"
DEPLOY_UI_ICON_MAGIC="✨"
DEPLOY_UI_PORT="8989"
DEPLOY_UI_SCHEME="http"
DEPLOY_UI_DASH_PATH="/__chimera/"
DEPLOY_UI_HEALTH_PATH="/__chimera/health"

# shellcheck source=deploy-ui.sh
source "$_DEPLOY_LIB_DIR/deploy-ui.sh"

chimera_build_metadata() {
    local repo_dir="$1"
    CHIMERA_GIT_VERSION=$(git -C "$repo_dir" describe --tags --always --dirty 2>/dev/null || echo 'dev')
    CHIMERA_GIT_COMMIT=$(git -C "$repo_dir" rev-parse --short HEAD 2>/dev/null || echo 'unknown')
    export CHIMERA_GIT_VERSION CHIMERA_GIT_COMMIT
}

chimera_parse_target() { deploy_ui_parse_target "$@"; }
chimera_deploy_state_file() { deploy_ui_deploy_state_file "$1"; }
chimera_save_deploy_last() {
    deploy_ui_save_deploy_last "$1" "$2" "$3" "$4" "${CHIMERA_GIT_VERSION:-}" "${CHIMERA_GIT_COMMIT:-}"
}
chimera_load_deploy_last() { deploy_ui_load_deploy_last "$1"; }
chimera_print_success() {
    deploy_ui_success "$1" "$2" "./scripts/deploy-remote.sh $1 --uninstall"
}

chimera_info()  { deploy_ui_info "$@"; }
chimera_warn()  { deploy_ui_warn "$@"; }
chimera_error() { deploy_ui_error "$@"; }
