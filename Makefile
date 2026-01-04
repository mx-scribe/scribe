.PHONY: help build build-ui build-go run test test-verbose test-coverage test-race clean serve dev lint fmt vet check-fmt ci install build-all

# Variables
BINARY_NAME=scribe
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags="-s -w -X github.com/mx-scribe/scribe/internal/version.BuildDate=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) -X github.com/mx-scribe/scribe/internal/version.GitCommit=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)"

# Default target
help:
	@echo "SCRIBE - Makefile Commands"
	@echo ""
	@echo "Build:"
	@echo "  make build          - Build binary with embedded UI"
	@echo "  make build-ui       - Build UI only"
	@echo "  make build-go       - Build Go binary only (no UI rebuild)"
	@echo "  make build-all      - Build for all platforms"
	@echo ""
	@echo "Development:"
	@echo "  make run            - Run with go run"
	@echo "  make serve          - Build and run server"
	@echo "  make dev            - Show dev mode instructions"
	@echo ""
	@echo "Testing:"
	@echo "  make test           - Run tests"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make test-race      - Run tests with race detector"
	@echo ""
	@echo "Quality:"
	@echo "  make lint           - Run linters"
	@echo "  make fmt            - Format code"
	@echo "  make vet            - Run go vet"
	@echo "  make check-fmt      - Check code formatting"
	@echo "  make ci             - Run all CI checks"
	@echo ""
	@echo "Other:"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make install        - Install binary to GOPATH/bin"

# Build Go binary with embedded UI
build: build-ui
	@mkdir -p bin
	@echo "🔨 Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/scribe
	@echo "✅ Build complete: bin/$(BINARY_NAME)"
	@ls -lh bin/$(BINARY_NAME)

# Build UI only
build-ui:
	@echo "🎨 Building UI..."
	cd web && bun run build
	@echo "✅ UI build complete"

# Build Go binary without UI rebuild
build-go:
	@mkdir -p bin
	@echo "🔨 Building $(BINARY_NAME) (Go only)..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/scribe
	@echo "✅ Build complete"

# Run with go run (development)
run:
	$(GO) run ./cmd/scribe

# Run tests
test:
	@echo "🧪 Running tests..."
	$(GO) test ./...

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running tests (verbose)..."
	$(GO) test -v ./...

# Run tests with coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	@./scripts/test-coverage.sh

# Run tests with race detector
test-race:
	@echo "🧪 Running tests with race detector..."
	$(GO) test -race ./...

# Run linters
lint:
	@echo "🔍 Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint not installed. Install with:"; \
		echo "    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Format code
fmt:
	@echo "🎨 Formatting code..."
	$(GO) fmt ./...
	@echo "✅ Code formatted"

# Run go vet
vet:
	@echo "🔍 Running go vet..."
	$(GO) vet ./...
	@echo "✅ go vet passed"

# Check if code is formatted
check-fmt:
	@echo "🔍 Checking code formatting..."
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "❌ The following files need formatting:"; \
		gofmt -s -l .; \
		exit 1; \
	else \
		echo "✅ All files are properly formatted"; \
	fi

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -rf bin/
	@rm -rf .build/
	@rm -rf web/dist/
	@echo "✅ Clean complete"

# Start server from built binary
serve: build
	@echo "🚀 Starting $(BINARY_NAME)..."
	./bin/$(BINARY_NAME) serve

# Development: run UI and Go server separately
dev:
	@echo "📝 Development mode:"
	@echo "   Terminal 1: cd web && bun run dev"
	@echo "   Terminal 2: go run ./cmd/scribe serve"

# Install binary to GOPATH/bin
install:
	@echo "📦 Installing $(BINARY_NAME)..."
	$(GO) install $(LDFLAGS) ./cmd/scribe
	@echo "✅ Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# Run all CI checks
ci: check-fmt vet test
	@echo "✅ All CI checks passed!"

# Build for multiple platforms
build-all: build-ui
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/scribe
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/scribe
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/scribe
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/scribe
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/scribe
	@echo "✅ All builds complete!"
	@ls -lh dist/

# Quick test (no cache)
test-nocache:
	@echo "🧪 Running tests (no cache)..."
	$(GO) test -count=1 ./...

# Run benchmarks
bench:
	@echo "⚡ Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...
