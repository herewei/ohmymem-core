# OhMyMem MCP Server Makefile
# Cross-platform compilation support

# Variables
BINARY_NAME := ohmymem
VERSION := v0.0.1
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(BUILD_DATE)

# Go build flags
GO_BUILD_FLAGS := -ldflags "$(LDFLAGS)" -trimpath

# Go cache (override to avoid permission issues)
GOCACHE ?= /tmp/go-build

# Output directories
DIST_DIR := dist

# Default target
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build:
	GOCACHE=$(GOCACHE) go build $(GO_BUILD_FLAGS) -o $(BINARY_NAME) .

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BINARY_NAME)
	rm -rf $(DIST_DIR)

# Cross-compilation targets
# Linux
.PHONY: build-linux
build-linux:
	GOCACHE=$(GOCACHE) GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOCACHE=$(GOCACHE) GOOS=linux GOARCH=arm64 go build $(GO_BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOCACHE=$(GOCACHE) GOOS=linux GOARCH=arm go build $(GO_BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm .

# macOS
.PHONY: build-darwin
build-darwin:
	GOCACHE=$(GOCACHE) GOOS=darwin GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOCACHE=$(GOCACHE) GOOS=darwin GOARCH=arm64 go build $(GO_BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .

# Windows
.PHONY: build-windows
build-windows:
	GOCACHE=$(GOCACHE) GOOS=windows GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# All platforms
.PHONY: build-all
build-all: clean
	mkdir -p $(DIST_DIR)
	$(MAKE) build-linux
	$(MAKE) build-darwin
	$(MAKE) build-windows
	@echo "Build complete. Output in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

# Run tests
.PHONY: test
test:
	GOCACHE=$(GOCACHE) go test ./...

# Run end-to-end tests (requires binary to be built)
.PHONY: test-e2e
test-e2e: build
	GOCACHE=$(GOCACHE) go test -v ./tests/e2e/...

# Run linter (requires golangci-lint)
.PHONY: lint
lint:
	golangci-lint run ./...

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Show help
.PHONY: help
help:
	@echo "OhMyMem MCP Server Build System"
	@echo ""
	@echo "Targets:"
	@echo "  all          - Build for current platform (default)"
	@echo "  build        - Build for current platform"
	@echo "  build-linux  - Build for Linux (amd64, arm64, arm)"
	@echo "  build-darwin - Build for macOS (amd64, arm64)"
	@echo "  build-windows- Build for Windows (amd64)"
	@echo "  build-all    - Build for all platforms"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  help         - Show this help"
