#!/bin/bash
set -e

REPO="LukasLow/smd-cli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and ARCH
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "Installing smd for ${OS}/${ARCH}..."

# Get latest release
BINARY="smd-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${BINARY}"

echo "Downloading from ${URL}..."
curl -fsSL "$URL" -o "/tmp/smd" || {
    echo "Failed to download. Please check if the release exists."
    exit 1
}

chmod +x "/tmp/smd"

# Install
echo "Installing to ${INSTALL_DIR}/smd..."
if [ -w "$INSTALL_DIR" ]; then
    mv "/tmp/smd" "$INSTALL_DIR/smd"
else
    sudo mv "/tmp/smd" "$INSTALL_DIR/smd"
fi

echo "smd installed successfully!"
echo "Run 'smd --version' to verify."
