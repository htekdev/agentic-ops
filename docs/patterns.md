# Agentic DevOps Patterns

A collection of proven workflow patterns for AI agent governance.

## Security Patterns

### Block Sensitive File Edits

Prevent agents from modifying files that contain secrets or credentials.

```yaml
name: Protect Secrets
blocking: true

on:
  file:
    paths:
      - '**/*.env*'
      - '**/secrets/**'
      - '**/*.pem'
      - '**/*.key'
      - '**/credentials*'
    types: [edit, create]

steps:
  - run: |
      echo "❌ Cannot modify sensitive file: ${{ event.file.path }}"
      echo "These files require manual review."
      exit 1
```

### Detect Hardcoded Secrets

Scan new code for potential credentials before they're committed.

```yaml
name: Secret Detection
blocking: true

on:
  file:
    types: [create, edit]
    paths: ['**/*.ts', '**/*.js', '**/*.py', '**/*.go']

steps:
  - name: Scan for secrets
    run: |
      patterns='(password|secret|api_key|apikey|token|credential)\s*[:=]'
      if grep -iE "$patterns" "${{ event.file.path }}"; then
        echo "❌ Potential hardcoded credentials detected"
        exit 1
      fi
```

### Block Dangerous Commands

Prevent agents from running destructive shell commands.

```yaml
name: Block Dangerous Commands
blocking: true

on:
  tool:
    name: [powershell, bash]

steps:
  - name: Check for dangerous patterns
    if: |
      contains(event.tool.args.command, 'rm -rf') ||
      contains(event.tool.args.command, 'DROP TABLE') ||
      contains(event.tool.args.command, 'format c:')
    run: |
      echo "❌ Dangerous command blocked"
      exit 1
```

## Quality Patterns

### Lint on Every Edit

Run linters immediately after file changes.

```yaml
name: Lint TypeScript
on:
  file:
    lifecycle: post
    paths: ['**/*.ts', '**/*.tsx']
    types: [edit]

blocking: false  # Report issues but don't block

steps:
  - run: npx eslint "${{ event.file.path }}" --fix
```

### Validate JSON/YAML Syntax

Catch syntax errors before they cause problems.

```yaml
name: Validate JSON
blocking: true

on:
  file:
    paths: ['**/*.json']
    types: [edit, create]

steps:
  - name: Check JSON syntax
    run: |
      if ! cat "${{ event.file.path }}" | jq . > /dev/null 2>&1; then
        echo "❌ Invalid JSON syntax"
        exit 1
      fi
      echo "✓ Valid JSON"
```

### Type Check on Save

Run TypeScript compiler after edits.

```yaml
name: Type Check
on:
  file:
    lifecycle: post
    paths: ['**/*.ts', '**/*.tsx']
    types: [edit]

blocking: false

steps:
  - run: npx tsc --noEmit
```

## Architecture Patterns

### Enforce Import Boundaries

Prevent imports that violate architectural rules.

```yaml
name: Import Boundaries
blocking: true

on:
  file:
    paths: ['src/domain/**/*.ts']
    types: [edit, create]

steps:
  - name: Check for infrastructure imports
    run: |
      if grep -E "from ['\"].*infrastructure" "${{ event.file.path }}"; then
        echo "❌ Domain layer cannot import from infrastructure"
        exit 1
      fi
```

### Require Tests with Source Changes

Ensure test files accompany source changes.

```yaml
name: Require Tests
blocking: true

on:
  commit:
    paths: ['src/**']
    paths-ignore: ['src/**/*.test.*', 'src/**/*.spec.*']

steps:
  - name: Check for test files
    run: |
      files="${{ toJSON(event.commit.files) }}"
      if ! echo "$files" | grep -qE '\.(test|spec)\.(ts|js)'; then
        echo "❌ Source changes require accompanying test files"
        exit 1
      fi
```

## Git Workflow Patterns

### Enforce Commit Message Format

Require conventional commit messages.

```yaml
name: Commit Message Format
blocking: true

on:
  commit: {}

steps:
  - name: Check format
    run: |
      msg="${{ event.commit.message }}"
      pattern='^(feat|fix|docs|style|refactor|test|chore)(\(.+\))?: .+'
      if ! echo "$msg" | grep -qE "$pattern"; then
        echo "❌ Commit message must follow conventional commits format"
        echo "   Example: feat(auth): add login endpoint"
        exit 1
      fi
```

### Protect Main Branch

Block direct commits to protected branches.

```yaml
name: Protect Main
blocking: true

on:
  commit:
    branches: [main, master, release/*]

steps:
  - run: |
      echo "❌ Direct commits to protected branches are not allowed"
      echo "   Please create a feature branch and open a PR"
      exit 1
```

## Notification Patterns

### Audit Trail

Log all agent actions for compliance.

```yaml
name: Audit Log
on:
  hooks:
    types: [postToolUse]

blocking: false

steps:
  - name: Log action
    run: |
      echo "[$(date -Iseconds)] Tool: ${{ event.tool.name }}, Path: ${{ event.file.path }}" >> .agent-audit.log
```

## Tips for Writing Effective Workflows

1. **Keep workflows focused** — One concern per workflow
2. **Use `blocking: false` for advisory checks** — Lint warnings shouldn't stop work
3. **Use `blocking: true` for security** — Never compromise on secrets protection
4. **Leverage `lifecycle: post`** — Validation after edit is less disruptive
5. **Test workflows with `gh hookflow test`** — Verify before committing
