# Agentic-Ops Copilot Instructions

This repository contains the `agentic-ops` CLI and Copilot plugin for local agent workflow execution.

## Architecture

```
agentic-ops/
├── cmd/agentic-ops/     # CLI entry point
├── internal/
│   ├── discover/        # Workflow file discovery
│   ├── schema/          # Workflow types and validation
│   ├── trigger/         # Event-to-trigger matching
│   ├── expression/      # ${{ }} expression engine
│   ├── runner/          # Step execution
│   └── concurrency/     # Semaphore for parallel control
├── packages/hooks/      # Copilot hook scripts
├── skills/              # SKILL.md for agent guidance
└── testdata/            # Test fixtures
```

## Development Workflow

1. Make changes to Go code
2. Run tests: `go test ./...`
3. Build: `go build -o bin/agentic-ops ./cmd/agentic-ops`
4. Test locally: `copilot plugin install .`

## Key Patterns

### Expression Engine
- Parser in `internal/expression/parser.go`
- Evaluator in `internal/expression/evaluator.go`
- Functions follow GitHub Actions parity

### Trigger Matching
- All trigger types in `internal/trigger/matcher.go`
- Glob patterns use `**` for recursive matching
- Negation with `!` prefix

### Hook Integration
- Scripts in `packages/hooks/scripts/`
- Both PowerShell and Bash versions
- Read event JSON from stdin, output decision JSON

## Testing

```bash
go test ./... -v              # All tests
go test ./internal/expression/... -v  # Specific package
go test ./... -coverprofile=coverage.out  # With coverage
```

## Cross-Compilation

```bash
GOOS=windows GOARCH=amd64 go build -o bin/agentic-ops-windows-amd64.exe ./cmd/agentic-ops
GOOS=darwin GOARCH=arm64 go build -o bin/agentic-ops-darwin-arm64 ./cmd/agentic-ops
GOOS=linux GOARCH=amd64 go build -o bin/agentic-ops-linux-amd64 ./cmd/agentic-ops
```
