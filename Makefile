.PHONY: build test coverage clean install

# Build the CLI
build:
	go build -o bin/agentic-ops ./cmd/agentic-ops

# Build for all platforms
build-all:
	GOOS=windows GOARCH=amd64 go build -o bin/agentic-ops-windows-amd64.exe ./cmd/agentic-ops
	GOOS=darwin GOARCH=amd64 go build -o bin/agentic-ops-darwin-amd64 ./cmd/agentic-ops
	GOOS=darwin GOARCH=arm64 go build -o bin/agentic-ops-darwin-arm64 ./cmd/agentic-ops
	GOOS=linux GOARCH=amd64 go build -o bin/agentic-ops-linux-amd64 ./cmd/agentic-ops

# Run tests
test:
	go test ./... -v

# Run tests with race detector
test-race:
	go test ./... -race -v

# Run tests with coverage
coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Show coverage in terminal
coverage-text:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Lint (requires golangci-lint)
lint:
	golangci-lint run

# Install CLI locally
install:
	go install ./cmd/agentic-ops

# Run the CLI
run:
	go run ./cmd/agentic-ops $(ARGS)
