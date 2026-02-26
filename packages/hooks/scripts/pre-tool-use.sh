#!/bin/bash
set -e

# Resolve plugin root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  darwin)
    if [ "$ARCH" = "arm64" ]; then
      BIN_NAME="agentic-ops-darwin-arm64"
    else
      BIN_NAME="agentic-ops-darwin-amd64"
    fi
    ;;
  linux)
    BIN_NAME="agentic-ops-linux-amd64"
    ;;
  *)
    # Unknown OS, allow by default
    echo '{"permissionDecision":"allow"}'
    exit 0
    ;;
esac

CLI="$PLUGIN_ROOT/bin/$BIN_NAME"

# Check if CLI exists
if [ ! -x "$CLI" ]; then
  echo '{"permissionDecision":"allow"}'
  exit 0
fi

# Read input from stdin
INPUT=$(cat)

if [ -z "$INPUT" ]; then
  echo '{"permissionDecision":"allow"}'
  exit 0
fi

# Extract fields using jq
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName // empty')
SESSION_CWD=$(echo "$INPUT" | jq -r '.cwd // empty')
TOOL_ARGS=$(echo "$INPUT" | jq -c '.toolArgs // {}')

# Build event JSON
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
EVENT=$(jq -n \
  --arg tool_name "$TOOL_NAME" \
  --argjson tool_args "$TOOL_ARGS" \
  --arg cwd "$SESSION_CWD" \
  --arg timestamp "$TIMESTAMP" \
  '{
    hook: {
      type: "preToolUse",
      tool: {
        name: $tool_name,
        args: $tool_args
      },
      cwd: $cwd
    },
    tool: {
      name: $tool_name,
      args: $tool_args,
      hook_type: "preToolUse"
    },
    cwd: $cwd,
    timestamp: $timestamp
  }')

# Run CLI
RESULT=$(echo "$EVENT" | "$CLI" run --event - --dir "$SESSION_CWD" 2>/dev/null) || true

# Output result or default to allow
if echo "$RESULT" | jq -e '.permissionDecision' > /dev/null 2>&1; then
  echo "$RESULT"
else
  echo '{"permissionDecision":"allow"}'
fi

exit 0
