# Agentic-Ops

Local workflow engine for agentic DevOps - execute GitHub Actions-like workflows triggered by Copilot agent hooks.

## Overview

Agentic-Ops enables governance for AI agents without sacrificing velocity. Define workflows that run locally in response to:

- **Agent hooks** (`preToolUse`, `postToolUse`)
- **File changes** (create, edit)
- **Git commits**
- **Git pushes**

Workflows use a familiar GitHub Actions-like syntax, making it easy to create automated checks, linting, security scans, and approval gates.

## Installation

```bash
# Install as a Copilot CLI plugin
copilot plugin install htekdev/agentic-ops
```

## Quick Start

Create `.github/agent-workflows/lint.yml`:

```yaml
name: Lint JavaScript
description: Run ESLint on JS file edits

on:
  file:
    types: [edit]
    paths: ['**/*.js']
    paths-ignore: ['node_modules/**']

steps:
  - name: Run ESLint
    run: npx eslint "${{ event.file.path }}"
    continue-on-error: true
```

## Workflow Syntax

### Triggers

```yaml
on:
  # Agent hook events
  hooks:
    types: [preToolUse, postToolUse]
    tools: [edit, create]

  # Specific tool with arg filtering
  tool:
    name: edit
    args:
      path: '**/*.env*'
    if: ${{ contains(event.tool.args.path, 'secrets') }}

  # File change events
  file:
    types: [create, edit]
    paths: ['src/**/*.ts']
    paths-ignore: ['**/*.test.ts']

  # Git commit events
  commit:
    paths: ['src/**']
    branches: [main, 'release/**']

  # Git push events
  push:
    branches: [main]
    tags: ['v*']
```

### Steps

```yaml
steps:
  - name: Step name
    if: ${{ condition }}
    run: echo "Hello ${{ event.file.path }}"
    shell: pwsh  # pwsh, bash, sh, cmd
    working-directory: ./src
    timeout: 30
    continue-on-error: false
    env:
      NODE_ENV: test
```

### Expressions

```yaml
# Context access
${{ event.file.path }}
${{ event.hook.tool.name }}
${{ env.NODE_ENV }}

# Functions
${{ contains(path, 'src') }}
${{ startsWith(path, 'src/') }}
${{ endsWith(path, '.ts') }}
${{ format('File: {0}', path) }}

# Operators
${{ a == b }}
${{ a && b || c }}
${{ !condition }}
```

## Configuration

### Blocking Mode

```yaml
blocking: true   # Default: stop agent on failure
blocking: false  # Log warning, allow agent to continue
```

### Concurrency Control

```yaml
concurrency:
  group: lint-${{ event.cwd }}
  max-parallel: 2
```

## CLI Usage

```bash
# Discover workflows
agentic-ops discover

# Validate workflows
agentic-ops validate

# Run workflows for an event
agentic-ops run --event '{"hook":{"type":"preToolUse",...}}'

# Run specific workflow
agentic-ops run --workflow lint --event '...'
```

## Examples

### Block Sensitive Files

```yaml
name: Block Env Edits
blocking: true

on:
  tool:
    name: edit
    args:
      path: '**/*.env*'

steps:
  - run: |
      echo "Cannot edit environment files"
      exit 1
```

### Security Scan

```yaml
name: Security Scan
on:
  file:
    types: [create]
    paths: ['**/*.js', '**/*.ts']

steps:
  - name: Check for secrets
    run: |
      if grep -E "(password|secret)" "${{ event.file.path }}"; then
        exit 1
      fi
```

## Development

```bash
# Build
go build -o bin/agentic-ops ./cmd/agentic-ops

# Test
go test ./...

# Test with coverage
go test ./... -coverprofile=coverage.out
```

## License

MIT
