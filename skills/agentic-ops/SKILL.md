---
name: agentic-ops
description: Expert skill for creating agent-workflows - local GitHub Actions-like workflows triggered by Copilot hooks. USE THIS SKILL when building governance workflows for agent behavior, creating preToolUse/postToolUse gates, or implementing automated checks. Trigger phrases include "create agent workflow", "add workflow gate", "lint on edit", "security check", "block sensitive files", "agent governance".
---

# Agentic-Ops: Agent Workflow Development

This skill provides comprehensive guidance for creating agent-workflows - local workflows that run in response to Copilot agent hooks, file changes, commits, and pushes.

## Quick Start

Create workflows in `.github/agent-workflows/<name>.yml`:

```yaml
name: Lint JavaScript
on:
  file:
    types: [edit]
    paths: ['**/*.js']

steps:
  - name: Run ESLint
    run: npx eslint "${{ event.file.path }}"
```

## Workflow File Location

```
.github/agent-workflows/
├── lint-on-edit.yml      # Lint files when edited
├── security-scan.yml     # Security checks
├── require-approval.yml  # Require approval for sensitive files
└── notify-changes.yml    # Post-edit notifications
```

## Workflow Schema

### Basic Structure

```yaml
name: Workflow Name              # Required: Display name
description: What this does      # Optional: Description

blocking: true                   # Default: true (deny stops agent)
                                 # false = log warning, allow action

concurrency:                     # Optional: Limit parallel execution
  group: lint-${{ event.cwd }}   # Group identifier (supports expressions)
  max-parallel: 2                # Max concurrent workflows in group

on:                              # Required: Triggers
  # ... trigger configuration

env:                             # Optional: Environment variables
  KEY: value

steps:                           # Required: Steps to execute
  - name: Step name
    run: echo "hello"
```

### Trigger Types

#### 1. Hook Triggers (Broad)

Match any agent hook event:

```yaml
on:
  hooks:
    types:
      - preToolUse      # Before tool executes
      - postToolUse     # After tool completes
    tools:              # Optional: Filter by tool name
      - edit
      - create
      - powershell
```

#### 2. Tool Triggers (Granular)

Match specific tools with argument filtering:

```yaml
on:
  tool:
    name: edit                           # Exact tool name
    args:
      path: '**/*.env*'                  # Glob pattern on args
    if: ${{ contains(event.tool.args.path, 'secrets') }}

  # Or multiple tools
  tools:
    - name: edit
      args:
        path: '**/*.env'
    - name: create
      args:
        path: '**/credentials*'
    - name: powershell
      if: ${{ contains(event.tool.args.command, 'rm -rf') }}
```

#### 3. File Triggers

Match file create/edit events:

```yaml
on:
  file:
    types:
      - create
      - edit
    paths:
      - 'src/**/*.ts'
      - '!src/**/*.test.ts'   # Negation with !
    paths-ignore:
      - 'node_modules/**'
      - '**/*.d.ts'
```

#### 4. Commit Triggers

Match git commit events:

```yaml
on:
  commit:
    paths:
      - 'src/**'
    paths-ignore:
      - '**/*.md'
    branches:
      - main
      - 'release/**'
    branches-ignore:
      - 'release/**-alpha'
```

#### 5. Push Triggers

Match git push events:

```yaml
on:
  push:
    branches:
      - main
      - 'feature/**'
    branches-ignore:
      - 'feature/**-wip'
    tags:
      - 'v*'
    tags-ignore:
      - '*-beta'
```

### Steps

```yaml
steps:
  - name: Step Name              # Optional: Display name
    if: ${{ expression }}        # Optional: Condition
    run: |                       # Shell command
      echo "Hello"
      npm run lint
    shell: pwsh                  # Optional: pwsh, bash, sh, cmd
    working-directory: ./src     # Optional: Working directory
    timeout: 30                  # Optional: Timeout in seconds
    continue-on-error: false     # Optional: Don't fail workflow on error
    env:                         # Optional: Step-level env vars
      NODE_ENV: test

  - name: Use Action
    uses: htekdev/action@v1      # Reusable action
    with:
      input1: value
```

## Expression Syntax

Expressions use `${{ }}` syntax with GitHub Actions parity:

### Context Objects

| Context | Description | Example |
|---------|-------------|---------|
| `event` | Event payload | `event.file.path` |
| `event.hook` | Hook details | `event.hook.type` |
| `event.tool` | Tool invocation | `event.tool.name` |
| `event.file` | File change | `event.file.content` |
| `event.commit` | Commit info | `event.commit.sha` |
| `event.push` | Push info | `event.push.ref` |
| `env` | Environment vars | `env.NODE_ENV` |
| `steps` | Step outputs | `steps.lint.outcome` |

### Operators

```yaml
# Comparison
${{ event.file.path == 'main.go' }}
${{ count > 5 }}

# Logical
${{ a && b }}
${{ a || b }}
${{ !condition }}

# Parentheses
${{ (a || b) && c }}
```

### Built-in Functions

| Function | Description | Example |
|----------|-------------|---------|
| `contains(search, item)` | Check if contains | `contains(path, 'src')` |
| `startsWith(str, prefix)` | String starts with | `startsWith(path, 'src/')` |
| `endsWith(str, suffix)` | String ends with | `endsWith(path, '.ts')` |
| `format(str, args...)` | String formatting | `format('Hello {0}', name)` |
| `join(array, sep)` | Join array | `join(labels, ', ')` |
| `toJSON(value)` | Convert to JSON | `toJSON(event)` |
| `fromJSON(str)` | Parse JSON | `fromJSON(data)` |
| `always()` | Always true | Used in `if:` |
| `success()` | Previous succeeded | Used in `if:` |
| `failure()` | Previous failed | Used in `if:` |

## Common Patterns

### 1. Lint on Edit

```yaml
name: Lint TypeScript
on:
  file:
    types: [edit]
    paths: ['**/*.ts', '**/*.tsx']
    paths-ignore: ['**/*.test.ts']

steps:
  - name: Run ESLint
    run: npx eslint "${{ event.file.path }}" --fix
    continue-on-error: true
```

### 2. Block Sensitive Files

```yaml
name: Block Env File Edits
blocking: true

on:
  tool:
    name: edit
    args:
      path: '**/*.env*'

steps:
  - name: Deny edit
    run: |
      echo "Cannot edit environment files directly"
      exit 1
```

### 3. Security Scan on Create

```yaml
name: Security Scan
on:
  file:
    types: [create]
    paths: ['**/*.js', '**/*.ts']

steps:
  - name: Check for secrets
    run: |
      if grep -E "(password|secret|api_key)" "${{ event.file.path }}"; then
        echo "Potential secret detected!"
        exit 1
      fi
```

### 4. Require Approval for Certain Paths

```yaml
name: Require Approval
blocking: true

on:
  tool:
    name: edit
    if: ${{ startsWith(event.tool.args.path, 'src/core/') }}

steps:
  - name: Check approval
    run: |
      # Integration with approval system
      if ! check-approval "${{ event.tool.args.path }}"; then
        echo "Editing core files requires approval"
        exit 1
      fi
```

### 5. Post-Edit Notifications

```yaml
name: Notify on Edit
blocking: false

on:
  hooks:
    types: [postToolUse]
    tools: [edit, create]

steps:
  - name: Log change
    run: |
      echo "File modified: ${{ event.tool.args.path }}"
      # Send to logging service
```

## Event Payload Reference

### Hook Event

```json
{
  "hook": {
    "type": "preToolUse",
    "tool": {
      "name": "edit",
      "args": { "path": "...", "old_str": "...", "new_str": "..." }
    },
    "cwd": "/project"
  },
  "cwd": "/project",
  "timestamp": "2026-02-26T12:00:00Z"
}
```

### File Event

```json
{
  "file": {
    "path": "src/main.ts",
    "action": "edit",
    "content": "..."
  },
  "cwd": "/project",
  "timestamp": "2026-02-26T12:00:00Z"
}
```

### Commit Event

```json
{
  "commit": {
    "sha": "abc123",
    "message": "feat: add feature",
    "author": "dev@example.com",
    "files": [
      { "path": "src/feature.ts", "status": "added" }
    ]
  },
  "cwd": "/project",
  "timestamp": "2026-02-26T12:00:00Z"
}
```

## CLI Commands

```bash
# Discover workflows
agentic-ops discover

# Validate workflows
agentic-ops validate

# Run workflows for an event
agentic-ops run --event '{"hook":...}'

# Run specific workflow
agentic-ops run --workflow lint --event '...'

# List trigger types
agentic-ops triggers
```

## Troubleshooting

### Workflow Not Triggering

1. Check trigger configuration matches event type
2. Verify path patterns are correct
3. Check if `paths-ignore` is excluding your file
4. Ensure workflow is in `.github/agent-workflows/`

### Step Failing

1. Check step `if:` condition
2. Verify shell command syntax
3. Check `timeout` isn't too short
4. Review step output for errors

### Expression Errors

1. Verify property path exists in event context
2. Check function argument count
3. Use `toJSON(event)` to debug event structure
