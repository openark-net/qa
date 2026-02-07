#!/bin/bash
set -e

GITHUB_REPO="openark-net/qa"
BINARY_NAME="qa"
INSTALL_DIR="/usr/local/bin"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "Fetching the latest release of $BINARY_NAME..."
LATEST_RELEASE=$(curl -sL "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo "Failed to fetch the latest release."
    exit 1
fi

echo "Latest release: $LATEST_RELEASE"

DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$LATEST_RELEASE/${BINARY_NAME}_${OS}_${ARCH}"

echo "Downloading $BINARY_NAME..."
curl -sL "$DOWNLOAD_URL" -o "$BINARY_NAME"

chmod +x "$BINARY_NAME"

echo "Installing $BINARY_NAME to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_NAME" "$INSTALL_DIR"
else
    sudo mv "$BINARY_NAME" "$INSTALL_DIR"
fi

if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    echo "$BINARY_NAME installed successfully to $INSTALL_DIR"
else
    echo "Installation failed."
    exit 1
fi
