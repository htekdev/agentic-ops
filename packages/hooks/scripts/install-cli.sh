#!/bin/bash
#
# Downloads the agentic-ops CLI binary from GitHub releases.
# This script is automatically run by the plugin hooks when the CLI is not found or needs updating.
#
# Usage:
#   install-cli.sh [version] [dest_dir] [--check-only]
#

set -e

VERSION="${1:-latest}"
DEST_DIR="${2:-}"
CHECK_ONLY="${3:-}"

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
VERSION_FILE="$DEST_DIR/.version"

# Determine download URL and get latest version
REPO_OWNER="htekdev"
REPO_NAME="agentic-ops-cli"

LATEST_VERSION=""
if [ "$VERSION" = "latest" ]; then
    API_URL="https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest"
    LATEST_VERSION=$(curl -sL --connect-timeout 5 "$API_URL" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$LATEST_VERSION" ]; then
        if [ "$CHECK_ONLY" = "--check-only" ]; then
            echo '{"updateAvailable":false,"error":"Failed to check for updates"}'
            exit 0
        fi
        echo "Failed to fetch latest release version"
        exit 1
    fi
    VERSION="$LATEST_VERSION"
fi

# Check current version
CURRENT_VERSION=""
if [ -f "$VERSION_FILE" ]; then
    CURRENT_VERSION=$(cat "$VERSION_FILE" | tr -d '[:space:]')
fi

# Check only mode
if [ "$CHECK_ONLY" = "--check-only" ]; then
    if [ "$CURRENT_VERSION" != "$LATEST_VERSION" ] && [ -n "$LATEST_VERSION" ]; then
        echo "{\"updateAvailable\":true,\"currentVersion\":\"$CURRENT_VERSION\",\"latestVersion\":\"$LATEST_VERSION\"}"
    else
        echo "{\"updateAvailable\":false,\"currentVersion\":\"$CURRENT_VERSION\",\"latestVersion\":\"$LATEST_VERSION\"}"
    fi
    exit 0
fi

DEST_PATH="$DEST_DIR/$BINARY_NAME"

# Skip if already up to date
if [ -f "$DEST_PATH" ] && [ "$CURRENT_VERSION" = "$VERSION" ]; then
    echo "agentic-ops CLI $VERSION is already installed"
    exit 0
fi

DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$VERSION/$BINARY_NAME"

echo "Downloading agentic-ops CLI $VERSION for $OS/$ARCH..."
echo "URL: $DOWNLOAD_URL"

curl -sL --connect-timeout 10 --max-time 60 "$DOWNLOAD_URL" -o "$DEST_PATH"
chmod +x "$DEST_PATH"

# Save version info
echo -n "$VERSION" > "$VERSION_FILE"

echo "Downloaded to: $DEST_PATH"

# Verify binary works
TEST_OUTPUT=$("$DEST_PATH" version 2>&1) || true
if [ $? -eq 0 ]; then
    echo "Verified CLI: $TEST_OUTPUT"
fi

echo "agentic-ops CLI installed successfully!"
