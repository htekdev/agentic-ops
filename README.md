# ⚠️ DEPRECATED - This repository has been renamed

> **This project has been rebranded to `hookflow` and moved to a new location.**

## New Repositories

| Purpose | New Location |
|---------|--------------|
| **CLI** | [htekdev/hookflow](https://github.com/htekdev/hookflow) |
| **Copilot Plugin** | [htekdev/hookflow-gh-copilot-plugin](https://github.com/htekdev/hookflow-gh-copilot-plugin) |

## Migration

If you have `agentic-ops` installed, uninstall and install the new plugin:

```bash
# Uninstall old plugin
copilot plugin uninstall htekdev/agentic-ops

# Install new plugin
copilot plugin install htekdev/hookflow-gh-copilot-plugin
```

## What Changed?

- **Name**: `agentic-ops` → `hookflow`
- **Workflow Directory**: `.github/agent-workflows/` → `.github/hooks/`
- **CLI Binary**: `agentic-ops` → `hookflow`

The syntax and functionality remain the same.

---

<details>
<summary>📜 Original README (archived)</summary>

# Agentic-Ops

A GitHub Copilot CLI plugin for local agent workflow governance. Execute GitHub Actions-like workflows triggered by agent hooks.

## Overview

Agentic-Ops enables governance for AI agents without sacrificing velocity. Define workflows that run locally in response to:
- **Agent hooks** (`preToolUse`, `postToolUse`)
- **File changes** (create, edit)
- **Git commits**
- **Git pushes**

Workflows use a familiar GitHub Actions-like syntax, making it easy to create automated checks, linting, security scans, and approval gates.

## License

MIT

</details>