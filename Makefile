# sing-box Distributed Management Platform Makefile

# Project information
PROJECT_NAME := sing-box-web
VERSION := $(shell git describe --tags --always --dirty)
BUILD_DATE := $(shell date +%Y-%m-%d\ %H:%M)
GO_VERSION := $(shell go version | awk '{print $$3}')

# Build configuration
GOOS ?= linux
GOARCH ?= amd64
CGO_ENABLED ?= 0

# Directory configuration
BUILD_DIR := build
PROTO_DIR := api
PROTO_OUT_DIR := pkg/pb
PKG_DIR := pkg

# Go build flags
LDFLAGS := -w -s \
	-X 'main.Version=$(VERSION)' \
	-X 'main.BuildDate=$(BUILD_DATE)' \
	-X 'main.GoVersion=$(GO_VERSION)'

# Proto files
PROTO_FILES := $(shell find $(PROTO_DIR) -name "*.proto")
PROTO_GO_FILES := $(patsubst $(PROTO_DIR)/%.proto,$(PROTO_OUT_DIR)/%.pb.go,$(PROTO_FILES))

# Binary targets
BINARIES := sing-box-web sing-box-api sing-box-agent

.PHONY: all build clean proto test lint fmt vet deps help

# Default target
all: clean proto build

help: ## Show help information
	@echo "Available Make targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { printf "  %-15s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# Build all binaries
build: deps proto ## Build all binaries
	@echo "Building all binaries..."
	@mkdir -p $(BUILD_DIR)
	@for binary in $(BINARIES); do \
		echo "Building $$binary..."; \
		CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$$binary \
		./cmd/$$binary; \
	done
	@echo "Build completed: $(BUILD_DIR)/"

# Build individual services
build-web: deps proto ## Build sing-box-web
	@echo "Building sing-box-web..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
	go build -ldflags "$(LDFLAGS)" \
	-o $(BUILD_DIR)/sing-box-web \
	./cmd/sing-box-web

build-api: deps proto ## Build sing-box-api
	@echo "Building sing-box-api..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
	go build -ldflags "$(LDFLAGS)" \
	-o $(BUILD_DIR)/sing-box-api \
	./cmd/sing-box-api

build-agent: deps proto ## Build sing-box-agent
	@echo "Building sing-box-agent..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
	go build -ldflags "$(LDFLAGS)" \
	-o $(BUILD_DIR)/sing-box-agent \
	./cmd/sing-box-agent

# Generate protobuf code
proto: $(PROTO_GO_FILES) ## Generate protobuf Go code

$(PROTO_OUT_DIR)/%.pb.go: $(PROTO_DIR)/%.proto
	@echo "Generating protobuf for $<..."
	@mkdir -p $(dir $@)
	@protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(PROTO_OUT_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT_DIR) \
		--go-grpc_opt=paths=source_relative \
		$<

# Clean protobuf generated files
clean-proto: ## Clean protobuf generated files
	@echo "Cleaning protobuf generated files..."
	@rm -rf $(PROTO_OUT_DIR)

# Dependency management
deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@go mod verify

# Code quality checks
test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Generate test coverage report
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run golangci-lint
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

# Code check collection
check: fmt vet lint test ## Run all code checks

# Clean build files
clean: ## Clean build files
	@echo "Cleaning build files..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Install development tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Docker related
docker-build: ## Build Docker images
	@echo "Building Docker images..."
	@docker build -t $(PROJECT_NAME):$(VERSION) .

docker-clean: ## Clean Docker images
	@echo "Cleaning Docker images..."
	@docker rmi $(PROJECT_NAME):$(VERSION) || true

# Development mode
dev-web: build-web ## Run web service in development mode
	@echo "Starting sing-box-web in development mode..."
	@./$(BUILD_DIR)/sing-box-web --config=configs/web.yaml --log-level=debug

dev-api: build-api ## Run api service in development mode
	@echo "Starting sing-box-api in development mode..."
	@./$(BUILD_DIR)/sing-box-api --config=configs/api.yaml --log-level=debug

dev-agent: build-agent ## Run agent service in development mode
	@echo "Starting sing-box-agent in development mode..."
	@./$(BUILD_DIR)/sing-box-agent --config=configs/agent.yaml --log-level=debug

# Production build
release: clean check build ## Production build
	@echo "Release build completed: $(VERSION)"

# Show build information
info: ## Show build information
	@echo "Project: $(PROJECT_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "GOOS: $(GOOS)"
	@echo "GOARCH: $(GOARCH)"
	@echo "CGO_ENABLED: $(CGO_ENABLED)"