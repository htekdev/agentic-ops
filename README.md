# Agentic DevOps

**Shift-left DevOps for AI agents.** Run governance, validation, and automation at the moment code is written — not after it hits CI/CD.

## The Problem

AI agents write code at machine speed. They can refactor seventeen files in ninety seconds. By the time your CI pipeline catches a bug, the agent has already moved on to the next hundred changes.

Traditional DevOps catches problems too late:

| Stage | Feedback Delay | Agent Has Already... |
|-------|----------------|---------------------|
| PR Review | Hours to days | Moved to next feature |
| CI Pipeline | Minutes | Changed 50 more files |
| Pre-commit hooks | Seconds | Made 10 more edits |

**Agentic DevOps** shifts validation to the exact moment of creation — before code even exists on disk.

## The Solution

Intercept agent actions in real-time using **hooks**. When an AI agent tries to edit a file, make a commit, or run a command, your workflows execute *before* or *after* the action:

```
Agent: "Edit .env file"
         │
         ▼
    preToolUse hook ──→ Workflow: "Block .env edits"
         │                        │
         │                    exit 1
         │                        │
         ▼                        ▼
      DENIED ◄────────────── Workflow fails
         │
    Agent receives
    feedback, adjusts
```

## Core Principles

### 1. Protect Velocity, Don't Restrict It

DevOps was invented to *enable* speed, not slow it down. Agentic DevOps is the same — let agents move fast, but catch mistakes instantly.

### 2. Shift Left to the Moment of Creation

Don't wait for CI. Don't wait for PR review. Validate at the exact moment code is being written.

### 3. Fail Fast, Self-Correct Faster

When a workflow blocks an action, the agent sees the failure immediately and can self-correct — all within the same session.

### 4. Use Familiar Syntax

If you can write a GitHub Actions workflow, you can write an agent governance workflow.

## Common Patterns

### Block Sensitive Files

```yaml
name: Protect Secrets
blocking: true

on:
  file:
    paths: ['**/*.env*', '**/secrets/**', '**/*.pem']
    types: [edit, create]

steps:
  - run: |
      echo "❌ Cannot modify sensitive files"
      exit 1
```

### Lint on Every Edit

```yaml
name: Lint TypeScript
on:
  file:
    lifecycle: post
    paths: ['**/*.ts']
    types: [edit]

steps:
  - run: npx eslint "${{ event.file.path }}" --fix
```

### Require Tests with Source Changes

```yaml
name: Require Tests
blocking: true

on:
  commit:
    paths: ['src/**']
    paths-ignore: ['src/**/*.test.*']

steps:
  - name: Check for test files
    run: |
      if ! echo "${{ event.commit.files }}" | grep -q '\.test\.'; then
        echo "❌ Source changes require accompanying tests"
        exit 1
      fi
```

### Security Scan New Files

```yaml
name: Secret Detection
blocking: true

on:
  file:
    types: [create]
    paths: ['**/*.ts', '**/*.js', '**/*.py']

steps:
  - run: |
      if grep -E "(password|secret|api_key)\s*=" "${{ event.file.path }}"; then
        echo "❌ Hardcoded credentials detected"
        exit 1
      fi
```

## Tools

### gh-hookflow (Recommended)

A GitHub CLI extension that implements agentic DevOps for GitHub Copilot.

```bash
# Install
gh extension install htekdev/gh-hookflow

# Initialize in your repo
gh hookflow init

# Create workflows in .github/hookflows/
```

**Features:**
- GitHub Actions-like YAML syntax
- Expression engine with `${{ }}` support
- Pre and post lifecycle hooks
- Blocking and non-blocking modes
- Cross-platform (Windows, macOS, Linux)

→ [gh-hookflow on GitHub](https://github.com/htekdev/gh-hookflow)

## Further Reading

- [Agentic DevOps: The Next Evolution of Shift Left](https://htek.dev/articles/agentic-devops-next-evolution-of-shift-left)
- [Agent Hooks: Controlling AI in Your Codebase](https://htek.dev/articles/agent-hooks-controlling-ai-codebase)
- [Test Enforcement Architecture for AI Agents](https://htek.dev/articles/test-enforcement-architecture-ai-agents)

## Contributing

Have a pattern that works well? Found a tool that implements agentic DevOps? Open a PR!

## License

MIT
