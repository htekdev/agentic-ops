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

# Normalize path to relative (if absolute and under cwd)
if echo "$TOOL_ARGS" | jq -e '.path' > /dev/null 2>&1; then
  ABS_PATH=$(echo "$TOOL_ARGS" | jq -r '.path')
  if [[ "$ABS_PATH" == /* ]] || [[ "$ABS_PATH" == "$SESSION_CWD"* ]]; then
    REL_PATH="${ABS_PATH#$SESSION_CWD}"
    REL_PATH="${REL_PATH#/}"
    if [ -n "$REL_PATH" ]; then
      TOOL_ARGS=$(echo "$TOOL_ARGS" | jq --arg path "$REL_PATH" '.path = $path')
    fi
  fi
fi

# Detect git commit/push commands
COMMIT_EVENT=""
PUSH_EVENT=""

if [[ "$TOOL_NAME" == "powershell" ]] || [[ "$TOOL_NAME" == "bash" ]] || [[ "$TOOL_NAME" == "shell" ]]; then
  COMMAND=$(echo "$TOOL_ARGS" | jq -r '.command // .script // .code // empty')
  
  if [ -n "$COMMAND" ]; then
    # Detect git commit
    if echo "$COMMAND" | grep -qE '\bgit\s+(commit|ci)\b'; then
      cd "$SESSION_CWD" 2>/dev/null || true
      
      # Get staged files
      STAGED_FILES="[]"
      if GIT_STATUS=$(git diff --cached --name-status 2>/dev/null); then
        STAGED_FILES=$(echo "$GIT_STATUS" | awk '
          BEGIN { printf "[" }
          NR > 1 { printf "," }
          {
            status = "modified"
            if ($1 == "A") status = "added"
            else if ($1 == "M") status = "modified"
            else if ($1 == "D") status = "deleted"
            else if ($1 == "R") status = "renamed"
            printf "{\"path\":\"%s\",\"status\":\"%s\"}", $2, status
          }
          END { printf "]" }
        ')
        if [ "$STAGED_FILES" = "[]" ] && [ -n "$GIT_STATUS" ]; then
          STAGED_FILES=$(echo "$GIT_STATUS" | while IFS=$'\t' read -r status file; do
            s="modified"
            case "$status" in
              A) s="added" ;;
              M) s="modified" ;;
              D) s="deleted" ;;
              R) s="renamed" ;;
            esac
            echo "{\"path\":\"$file\",\"status\":\"$s\"}"
          done | jq -s '.')
        fi
      fi
      
      # Get commit message from command
      MESSAGE=""
      if MSG=$(echo "$COMMAND" | grep -oP '(?<=-m\s['\''"])[^'\''"]+(?=['\''"])' 2>/dev/null); then
        MESSAGE="$MSG"
      elif MSG=$(echo "$COMMAND" | grep -oP '(?<=-m\s)\S+' 2>/dev/null); then
        MESSAGE="$MSG"
      fi
      
      # Get branch and author
      BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")
      AUTHOR=$(git config user.email 2>/dev/null || echo "")
      
      COMMIT_EVENT=$(jq -n \
        --arg sha "pending" \
        --arg message "$MESSAGE" \
        --arg author "$AUTHOR" \
        --arg branch "$BRANCH" \
        --argjson files "$STAGED_FILES" \
        '{sha: $sha, message: $message, author: $author, branch: $branch, files: $files}')
    fi
    
    # Detect git push
    if echo "$COMMAND" | grep -qE '\bgit\s+push\b'; then
      cd "$SESSION_CWD" 2>/dev/null || true
      
      BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")
      REF="refs/heads/$BRANCH"
      
      # Check for tag push
      if echo "$COMMAND" | grep -qE '\bgit\s+push\s+.*--tags\b'; then
        REF="refs/tags/latest"
      elif TAG=$(echo "$COMMAND" | grep -oP '(?<=git\s+push\s+origin\s+)(v[\d\.]+)' 2>/dev/null); then
        REF="refs/tags/$TAG"
      fi
      
      CURRENT_SHA=$(git rev-parse HEAD 2>/dev/null || echo "0000000")
      
      PUSH_EVENT=$(jq -n \
        --arg ref "$REF" \
        --arg before "0000000000000000000000000000000000000000" \
        --arg after "$CURRENT_SHA" \
        '{ref: $ref, before: $before, after: $after}')
    fi
  fi
fi

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

# Add commit event if detected
if [ -n "$COMMIT_EVENT" ]; then
  EVENT=$(echo "$EVENT" | jq --argjson commit "$COMMIT_EVENT" '. + {commit: $commit}')
fi

# Add push event if detected
if [ -n "$PUSH_EVENT" ]; then
  EVENT=$(echo "$EVENT" | jq --argjson push "$PUSH_EVENT" '. + {push: $push}')
fi

# Run CLI
RESULT=$(echo "$EVENT" | "$CLI" run --event - --dir "$SESSION_CWD" 2>/dev/null) || true

# Output result or default to allow
if echo "$RESULT" | jq -e '.permissionDecision' > /dev/null 2>&1; then
  echo "$RESULT"
else
  echo '{"permissionDecision":"allow"}'
fi

exit 0
