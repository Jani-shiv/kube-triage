#!/bin/bash
# install.sh - Install kube-triage on Linux/macOS

set -e

REPO="Jani-shiv/kube-triage"
BINARY="kube-triage"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

detect_os_arch() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) log_error "Unsupported architecture: $ARCH"; exit 1 ;;
    esac

    case $OS in
        linux|darwin) ;;
        *) log_error "Unsupported OS: $OS"; exit 1 ;;
    esac

    echo "${OS}_${ARCH}"
}

get_latest_version() {
    curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
}

download_and_install() {
    local version=$1
    local os_arch=$2
    local tmp_dir=$(mktemp -d)
    local archive_name=""

    if [[ "$os_arch" == *"windows"* ]]; then
        archive_name="${BINARY}_${version}_${os_arch}.zip"
    else
        archive_name="${BINARY}_${version}_${os_arch}.tar.gz"
    fi

    local download_url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"

    log_info "Downloading ${archive_name}..."
    curl -sL "${download_url}" -o "${tmp_dir}/${archive_name}"

    log_info "Extracting..."
    if [[ "$archive_name" == *.zip ]]; then
        unzip -q "${tmp_dir}/${archive_name}" -d "${tmp_dir}"
    else
        tar -xzf "${tmp_dir}/${archive_name}" -C "${tmp_dir}"
    fi

    log_info "Installing to ${INSTALL_DIR}..."
    sudo mv "${tmp_dir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    sudo chmod +x "${INSTALL_DIR}/${BINARY}"

    rm -rf "${tmp_dir}"
    log_success "Installed ${BINARY} ${version} to ${INSTALL_DIR}"
}

main() {
    log_info "Installing kube-triage..."

    if [[ $EUID -eq 0 ]] && [[ "$INSTALL_DIR" == "/usr/local/bin" ]]; then
        log_warn "Running as root. Consider using a non-root install directory."
    fi

    local os_arch=$(detect_os_arch)
    log_info "Detected OS/Arch: ${os_arch}"

    local version=$(get_latest_version)
    if [[ -z "$version" ]]; then
        log_error "Failed to get latest version"
        exit 1
    fi

    log_info "Latest version: ${version}"
    download_and_install "$version" "$os_arch"

    log_success "Installation complete!"
    echo ""
    echo "Run 'kube-triage --help' to get started."
    echo ""
    echo "Quick start:"
    echo "  kube-triage analyze pod <pod-name> -n <namespace>"
    echo ""
    echo "For AI-powered analysis, set one of these environment variables:"
    echo "  export NVIDIA_API_KEY=nvapi-xxxxx     # NVIDIA Nemotron (free tier)"
    echo "  export OLLAMA_HOST=http://localhost:11434  # Local Ollama"
    echo "  export OPENAI_API_KEY=sk-xxxxx        # OpenAI"
    echo "  export ANTHROPIC_API_KEY=sk-ant-xxxxx # Anthropic"
}

main "$@"