#!/bin/bash
# Krakenv Installer
# Usage: curl -sSL https://raw.githubusercontent.com/theburrowhub/krakenv/main/install.sh | bash
#
# Environment variables:
#   KRAKENV_INSTALL_DIR - Installation directory (default: /usr/local/bin)
#   KRAKENV_VERSION     - Specific version to install (default: latest)

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
GITHUB_REPO="theburrowhub/krakenv"
BINARY_NAME="krakenv"
INSTALL_DIR="${KRAKENV_INSTALL_DIR:-/usr/local/bin}"

# Banner
echo -e "${PURPLE}"
cat << 'EOF'
    🐙 KRAKENV INSTALLER
    When envs get complex, release the krakenv
EOF
echo -e "${NC}"

# Functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect OS
detect_os() {
    local os
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$os" in
        linux*)  echo "linux" ;;
        darwin*) echo "darwin" ;;
        mingw*|msys*|cygwin*) echo "windows" ;;
        *) error "Unsupported operating system: $os" ;;
    esac
}

# Detect architecture
detect_arch() {
    local arch
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        armv7l|armv6l) echo "arm" ;;
        i386|i686) echo "386" ;;
        *) error "Unsupported architecture: $arch" ;;
    esac
}

# Get latest version from GitHub
get_latest_version() {
    local version
    version=$(curl -sL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$version" ]; then
        error "Failed to fetch latest version. Check your internet connection."
    fi
    echo "$version"
}

# Download and install
install_krakenv() {
    local os="$1"
    local arch="$2"
    local version="$3"

    # Build download URL
    local filename="${BINARY_NAME}_${os}_${arch}"
    if [ "$os" = "windows" ]; then
        filename="${filename}.exe"
    fi
    
    # GoReleaser format: krakenv_Linux_x86_64.tar.gz
    local os_name
    local arch_name
    case "$os" in
        linux) os_name="Linux" ;;
        darwin) os_name="Darwin" ;;
        windows) os_name="Windows" ;;
    esac
    case "$arch" in
        amd64) arch_name="x86_64" ;;
        arm64) arch_name="arm64" ;;
        arm) arch_name="armv7" ;;
        386) arch_name="i386" ;;
    esac

    local archive_name="${BINARY_NAME}_${os_name}_${arch_name}"
    local archive_ext="tar.gz"
    if [ "$os" = "windows" ]; then
        archive_ext="zip"
    fi

    local download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${archive_name}.${archive_ext}"

    info "Downloading krakenv ${version} for ${os}/${arch}..."
    info "URL: ${download_url}"

    # Create temp directory
    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT

    # Download
    local archive_path="${tmp_dir}/${archive_name}.${archive_ext}"
    if ! curl -fsSL "$download_url" -o "$archive_path"; then
        error "Failed to download krakenv. Please check if version ${version} exists."
    fi

    # Extract
    info "Extracting..."
    cd "$tmp_dir"
    if [ "$archive_ext" = "tar.gz" ]; then
        tar -xzf "$archive_path"
    else
        unzip -q "$archive_path"
    fi

    # Find binary
    local binary_path
    binary_path=$(find "$tmp_dir" -name "$BINARY_NAME" -o -name "${BINARY_NAME}.exe" 2>/dev/null | head -1)
    if [ -z "$binary_path" ]; then
        error "Binary not found in archive"
    fi

    # Install
    info "Installing to ${INSTALL_DIR}..."
    
    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        cp "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        warn "Need sudo to install to ${INSTALL_DIR}"
        sudo cp "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    success "Installed krakenv to ${INSTALL_DIR}/${BINARY_NAME}"
}

# Verify installation
verify_installation() {
    if command -v krakenv &> /dev/null; then
        success "Installation verified!"
        echo ""
        krakenv version
        echo ""
        echo -e "${GREEN}🐙 Krakenv is ready!${NC}"
        echo ""
        echo "Quick start:"
        echo "  krakenv init           # Create a new .env.dist"
        echo "  krakenv generate .env  # Generate environment file"
        echo "  krakenv --help         # Show all commands"
        echo ""
    else
        warn "krakenv was installed but is not in your PATH"
        echo "Add ${INSTALL_DIR} to your PATH or run:"
        echo "  ${INSTALL_DIR}/krakenv --help"
    fi
}

# Main
main() {
    info "Detecting system..."
    
    local os
    os=$(detect_os)
    success "OS: $os"
    
    local arch
    arch=$(detect_arch)
    success "Architecture: $arch"
    
    local version="${KRAKENV_VERSION:-}"
    if [ -z "$version" ]; then
        info "Fetching latest version..."
        version=$(get_latest_version)
    fi
    success "Version: $version"
    
    echo ""
    install_krakenv "$os" "$arch" "$version"
    echo ""
    verify_installation
}

main "$@"

