# Agentic DevOps - Copilot Instructions

This repository is the **conceptual hub** for Agentic DevOps — shift-left DevOps for AI agents.

## Repository Structure

```
agentic-ops/
├── README.md           # Main manifesto and overview
├── docs/
│   ├── why.md          # The problem and philosophy
│   ├── patterns.md     # Common workflow patterns
│   └── getting-started.md  # Quick start guide
├── examples/           # Copy-paste workflow examples
│   ├── protect-secrets.yml
│   ├── lint-typescript.yml
│   ├── require-tests.yml
│   ├── secret-detection.yml
│   └── validate-json.yml
└── LICENSE
```

## Purpose

This is a **documentation and patterns repository**, not a tool implementation.

- Explains the "why" of Agentic DevOps
- Provides reusable workflow patterns
- Points to tools like gh-hookflow for implementation

## Related Repositories

- [gh-hookflow](https://github.com/htekdev/gh-hookflow) — The CLI tool that implements agentic workflows

## Contributing

- Add new patterns to `docs/patterns.md`
- Add example workflows to `examples/`
- Improve documentation clarity
