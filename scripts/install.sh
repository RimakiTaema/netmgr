#!/bin/bash
set -euo pipefail

# NetMgr Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/netmgr/netmgr/main/scripts/install.sh | bash

REPO="netmgr/netmgr"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="netmgr"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="aarch64" ;;
    *) echo "❌ Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux) 
        PLATFORM="linux"
        BINARY_EXT=""
        ;;
    darwin) 
        PLATFORM="macos"
        BINARY_EXT=""
        ;;
    mingw*|cygwin*|msys*)
        PLATFORM="windows"
        BINARY_EXT=".exe"
        ;;
    *) echo "❌ Unsupported OS: $OS"; exit 1 ;;
esac

echo "🚀 Installing NetMgr for $PLATFORM-$ARCH"

# Get latest release
echo "📡 Fetching latest release..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo "❌ Failed to get latest release"
    exit 1
fi

echo "📦 Latest version: $LATEST_RELEASE"

# Download binary
BINARY_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/netmgr-$PLATFORM-$ARCH$BINARY_EXT"
TEMP_FILE="/tmp/netmgr$BINARY_EXT"

echo "⬇️  Downloading from: $BINARY_URL"
if ! curl -L -o "$TEMP_FILE" "$BINARY_URL"; then
    echo "❌ Failed to download binary"
    exit 1
fi

# Install binary
echo "📥 Installing to $INSTALL_DIR"
if [ "$OS" != "mingw" ] && [ "$OS" != "cygwin" ] && [ "$OS" != "msys" ]; then
    sudo mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    # Set capabilities on Linux
    if [ "$PLATFORM" = "linux" ] && command -v setcap >/dev/null 2>&1; then
        sudo setcap 'cap_net_admin,cap_net_raw+ep' "$INSTALL_DIR/$BINARY_NAME" || true
    fi
else
    mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME$BINARY_EXT"
fi

# Verify installation
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    echo "✅ NetMgr installed successfully!"
    echo "🔧 Version: $($BINARY_NAME --version 2>/dev/null || echo $LATEST_RELEASE)"
    echo "📚 Run '$BINARY_NAME --help' to get started"
    
    if [ "$PLATFORM" = "linux" ] || [ "$PLATFORM" = "macos" ]; then
        echo "⚠️  Note: Most operations require root privileges"
        echo "   Use: sudo $BINARY_NAME <command>"
    fi
else
    echo "❌ Installation failed - binary not found in PATH"
    exit 1
fi
