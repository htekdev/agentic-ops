# Getting Started with Agentic DevOps

This guide walks you through setting up agentic governance for your AI coding sessions using [gh-hookflow](https://github.com/htekdev/gh-hookflow).

## Prerequisites

- [GitHub CLI](https://cli.github.com/) (`gh`) installed and authenticated
- [PowerShell Core](https://github.com/PowerShell/PowerShell) (`pwsh`) installed

## Installation

Install gh-hookflow as a GitHub CLI extension:

```bash
gh extension install htekdev/gh-hookflow
```

Verify the installation:

```bash
gh hookflow version
```

## Initialize Your Repository

Navigate to your project and run:

```bash
cd your-project
gh hookflow init
```

This creates:
- `.github/hookflows/` — Directory for your workflow files
- `.github/hooks/hooks.json` — Copilot CLI hook configuration
- `.github/hookflows/example.yml` — A starter workflow

## Your First Workflow

Let's create a workflow that blocks edits to `.env` files.

Create `.github/hookflows/protect-env.yml`:

```yaml
name: Protect Environment Files
description: Block edits to .env files

on:
  file:
    paths:
      - '**/*.env*'
    types:
      - edit
      - create

blocking: true

steps:
  - name: Block sensitive file
    run: |
      echo "❌ Cannot modify environment files"
      echo "File: ${{ event.file.path }}"
      exit 1
```

## Test Your Workflow

Before committing, test that it works:

```bash
gh hookflow test --event file --action edit --path ".env"
```

You should see the workflow trigger and block the edit.

## How It Works

When you use GitHub Copilot CLI with hookflow initialized:

1. **Agent requests an action** — "Edit the .env file"
2. **Copilot calls preToolUse hook** — hookflow receives the event
3. **hookflow matches workflows** — Finds `protect-env.yml` matches
4. **Workflow runs** — Executes steps, exits with code 1
5. **hookflow returns deny** — Tells Copilot to block the action
6. **Agent sees feedback** — Can self-correct and try something else

## Workflow Lifecycle

Workflows can run at two points:

### Pre (Default)
Runs **before** the action. Can block/deny.

```yaml
on:
  file:
    lifecycle: pre  # This is the default
    paths: ['**/*.env']
```

### Post
Runs **after** the action completes. For validation, linting.

```yaml
on:
  file:
    lifecycle: post
    paths: ['**/*.ts']
```

## Expression Syntax

Use `${{ }}` to access event data:

```yaml
steps:
  - name: Show file info
    run: |
      echo "File: ${{ event.file.path }}"
      echo "Action: ${{ event.file.action }}"
```

### Available Context

| Expression | Description |
|------------|-------------|
| `event.file.path` | Path of the file |
| `event.file.action` | Action: edit, create, delete |
| `event.tool.name` | Tool being called |
| `event.tool.args.*` | Tool arguments |
| `event.commit.message` | Commit message |
| `event.lifecycle` | pre or post |
| `env.VAR_NAME` | Environment variable |

### Functions

| Function | Example |
|----------|---------|
| `contains(str, substr)` | `contains(event.file.path, '.env')` |
| `startsWith(str, prefix)` | `startsWith(event.file.path, 'src/')` |
| `endsWith(str, suffix)` | `endsWith(event.file.path, '.ts')` |

## Common Commands

```bash
# List discovered workflows
gh hookflow discover

# Validate workflow syntax
gh hookflow validate

# Test with mock events
gh hookflow test --event file --action edit --path "src/index.ts"
gh hookflow test --event commit --path "src/index.ts"

# View logs (for debugging)
gh hookflow logs
gh hookflow logs -f  # Follow mode
```

## Next Steps

1. **Add more workflows** — See [patterns.md](patterns.md) for ideas
2. **Commit your workflows** — `git add .github/ && git commit -m "Add hookflow workflows"`
3. **Share with team** — Team members need hookflow installed:
   ```bash
   gh extension install htekdev/gh-hookflow
   ```

## Troubleshooting

### Workflow Not Triggering

1. Check the trigger matches your event type (`file`, `commit`, `tool`)
2. Verify path patterns use glob syntax (`**/*.ts`, not `*.ts`)
3. Ensure `types` includes the action (`edit`, `create`)
4. Check `lifecycle` matches hook type (`pre` for preToolUse)

### Enable Debug Logging

```bash
export HOOKFLOW_DEBUG=1
gh hookflow logs -f
```

### Validate Syntax

```bash
gh hookflow validate
```
