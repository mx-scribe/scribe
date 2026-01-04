#!/bin/bash
# SCRIBE Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/mx-scribe/scribe/main/install.sh | bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="mx-scribe/scribe"
INSTALL_DIR="${SCRIBE_INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="scribe"

echo ""
echo "  ███████╗ ██████╗██████╗ ██╗██████╗ ███████╗"
echo "  ██╔════╝██╔════╝██╔══██╗██║██╔══██╗██╔════╝"
echo "  ███████╗██║     ██████╔╝██║██████╔╝█████╗  "
echo "  ╚════██║██║     ██╔══██╗██║██╔══██╗██╔══╝  "
echo "  ███████║╚██████╗██║  ██║██║██████╔╝███████╗"
echo "  ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝╚═════╝ ╚══════╝"
echo ""
echo "  Smart logging for humans."
echo ""

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"

    case "$OS" in
        linux*)   OS="linux" ;;
        darwin*)  OS="darwin" ;;
        mingw*|msys*|cygwin*) OS="windows" ;;
        *)
            echo -e "${RED}Error: Unsupported operating system: $OS${NC}"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *)
            echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    echo -e "${GREEN}Detected platform:${NC} $PLATFORM"
}

# Get latest release version
get_latest_version() {
    echo -e "${GREEN}Fetching latest version...${NC}"

    if command -v curl &> /dev/null; then
        VERSION=$(curl -sI "https://github.com/$REPO/releases/latest" | grep -i "location:" | sed 's/.*\/tag\/v//' | tr -d '\r\n')
    elif command -v wget &> /dev/null; then
        VERSION=$(wget -qO- --server-response "https://github.com/$REPO/releases/latest" 2>&1 | grep -i "location:" | sed 's/.*\/tag\/v//' | tr -d '\r\n')
    else
        echo -e "${RED}Error: curl or wget required${NC}"
        exit 1
    fi

    if [ -z "$VERSION" ]; then
        echo -e "${RED}Error: Could not determine latest version${NC}"
        exit 1
    fi

    echo -e "${GREEN}Latest version:${NC} v$VERSION"
}

# Download binary
download_binary() {
    DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/scribe-${PLATFORM}.tar.gz"
    TMP_DIR=$(mktemp -d)
    TMP_FILE="$TMP_DIR/scribe.tar.gz"

    echo -e "${GREEN}Downloading:${NC} $DOWNLOAD_URL"

    if command -v curl &> /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"
    else
        wget -q "$DOWNLOAD_URL" -O "$TMP_FILE"
    fi

    echo -e "${GREEN}Extracting...${NC}"
    tar -xzf "$TMP_FILE" -C "$TMP_DIR"

    # Find the binary (could be scribe or scribe-VERSION-platform)
    BINARY=$(find "$TMP_DIR" -name "scribe*" -type f ! -name "*.tar.gz" | head -1)

    if [ -z "$BINARY" ]; then
        echo -e "${RED}Error: Binary not found in archive${NC}"
        exit 1
    fi

    chmod +x "$BINARY"
}

# Install binary
install_binary() {
    echo -e "${GREEN}Installing to:${NC} $INSTALL_DIR/$BINARY_NAME"

    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv "$BINARY" "$INSTALL_DIR/$BINARY_NAME"
    else
        echo -e "${YELLOW}Note: Requesting sudo access to install to $INSTALL_DIR${NC}"
        sudo mv "$BINARY" "$INSTALL_DIR/$BINARY_NAME"
    fi

    # Cleanup
    rm -rf "$TMP_DIR"
}

# Verify installation
verify_installation() {
    if command -v scribe &> /dev/null; then
        echo ""
        echo -e "${GREEN}✓ SCRIBE installed successfully!${NC}"
        echo ""
        scribe version
        echo ""
        echo "Quick start:"
        echo "  scribe serve              # Start server on :8080"
        echo "  scribe log 'Hello world'  # Send a log"
        echo ""
        echo "Visit http://localhost:8080 after starting the server."
        echo ""
    else
        echo ""
        echo -e "${YELLOW}Warning: scribe not found in PATH${NC}"
        echo "You may need to add $INSTALL_DIR to your PATH"
        echo ""
        echo "Add this to your ~/.bashrc or ~/.zshrc:"
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
    fi
}

# Main
main() {
    detect_platform
    get_latest_version
    download_binary
    install_binary
    verify_installation
}

main
