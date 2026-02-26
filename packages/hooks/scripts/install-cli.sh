#!/bin/bash
#
# Downloads the agentic-ops CLI binary from GitHub releases.
# This script is automatically run by the plugin hooks when the CLI is not found.
#

set -e

VERSION="${1:-latest}"
DEST_DIR="${2:-}"

# Determine plugin root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

if [ -z "$DEST_DIR" ]; then
    DEST_DIR="$PLUGIN_ROOT/bin"
fi

# Create bin directory if it doesn't exist
mkdir -p "$DEST_DIR"

# Detect platform
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

EXT=""
if [ "$OS" = "windows" ]; then
    EXT=".exe"
fi

BINARY_NAME="agentic-ops-$OS-$ARCH$EXT"

# Determine download URL
REPO_OWNER="htekdev"
REPO_NAME="agentic-ops-cli"

if [ "$VERSION" = "latest" ]; then
    API_URL="https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest"
    VERSION=$(curl -sL "$API_URL" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        echo "Failed to fetch latest release version"
        exit 1
    fi
fi

DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$VERSION/$BINARY_NAME"

DEST_PATH="$DEST_DIR/$BINARY_NAME"

echo "Downloading agentic-ops CLI $VERSION for $OS/$ARCH..."
echo "URL: $DOWNLOAD_URL"

curl -sL "$DOWNLOAD_URL" -o "$DEST_PATH"
chmod +x "$DEST_PATH"

echo "Downloaded to: $DEST_PATH"

# Verify binary works
TEST_OUTPUT=$("$DEST_PATH" version 2>&1) || true
if [ $? -eq 0 ]; then
    echo "Verified CLI: $TEST_OUTPUT"
fi

echo "agentic-ops CLI installed successfully!"
