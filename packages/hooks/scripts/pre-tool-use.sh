#!/bin/bash
set -e

# Resolve plugin root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
BIN_DIR="$PLUGIN_ROOT/bin"

# Priority 1: Check if agentic-ops is installed globally (in PATH)
CLI=""
if command -v agentic-ops &> /dev/null; then
    CLI="agentic-ops"
fi

# Priority 2: Fall back to bundled binary
if [ -z "$CLI" ]; then
    # Detect OS and architecture
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"

    case "$OS" in
      darwin)
        case "$ARCH" in
          arm64|aarch64) BIN_NAME="agentic-ops-darwin-arm64" ;;
          *) BIN_NAME="agentic-ops-darwin-amd64" ;;
        esac
        ;;
      linux)
        case "$ARCH" in
          arm64|aarch64) BIN_NAME="agentic-ops-linux-arm64" ;;
          *) BIN_NAME="agentic-ops-linux-amd64" ;;
        esac
        ;;
      *)
        # Unknown OS, allow by default
        echo '{"permissionDecision":"allow"}'
        exit 0
        ;;
    esac

    BUNDLED_CLI="$BIN_DIR/$BIN_NAME"
    INSTALL_SCRIPT="$SCRIPT_DIR/install-cli.sh"

    # Check if bundled CLI exists, auto-install if missing
    if [ ! -x "$BUNDLED_CLI" ]; then
      if [ -x "$INSTALL_SCRIPT" ]; then
        "$INSTALL_SCRIPT" "latest" "$BIN_DIR" 2>/dev/null || true
      fi
      
      # Check again after install
      if [ ! -x "$BUNDLED_CLI" ]; then
        echo '{"permissionDecision":"allow"}'
        exit 0
      fi
    else
      # CLI exists - check for updates periodically (once per hour)
      LAST_CHECK_FILE="$BIN_DIR/.last-update-check"
      SHOULD_CHECK=true
      
      if [ -f "$LAST_CHECK_FILE" ]; then
        LAST_CHECK=$(stat -c %Y "$LAST_CHECK_FILE" 2>/dev/null || stat -f %m "$LAST_CHECK_FILE" 2>/dev/null || echo 0)
        NOW=$(date +%s)
        HOURS_SINCE=$(( (NOW - LAST_CHECK) / 3600 ))
        if [ "$HOURS_SINCE" -lt 1 ]; then
          SHOULD_CHECK=false
        fi
      fi
      
      if [ "$SHOULD_CHECK" = true ]; then
        # Update timestamp first
        touch "$LAST_CHECK_FILE"
        
        # Check for updates in background (don't block the hook)
        ("$INSTALL_SCRIPT" "latest" "$BIN_DIR" 2>/dev/null &) || true
      fi
    fi
    
    CLI="$BUNDLED_CLI"
fi

# Read input from stdin
INPUT=$(cat)

if [ -z "$INPUT" ]; then
  echo '{"permissionDecision":"allow"}'
  exit 0
fi

# Extract cwd for CLI directory flag
SESSION_CWD=$(echo "$INPUT" | jq -r '.cwd // empty')

# Pass raw input directly to CLI with --raw flag
# The CLI will detect event types (git commit, push, file changes, etc.)
RESULT=$(echo "$INPUT" | "$CLI" run --raw --dir "$SESSION_CWD" 2>/dev/null) || true

# Output result or default to allow
if echo "$RESULT" | jq -e '.permissionDecision' > /dev/null 2>&1; then
  echo "$RESULT"
else
  echo '{"permissionDecision":"allow"}'
fi

exit 0
