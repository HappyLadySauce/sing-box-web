.PHONY: all build build-no-tracing test clean generate lint install-tools install-skywalking check-skywalking docker-build help

# 构建变量
VERSION ?= $(shell git describe --tags --dirty --always 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Go 变量
GO_VERSION := 1.21
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# 输出目录
BUILD_DIR := build
DIST_DIR := dist

# 编译标志
LDFLAGS := -X 'sing-box-web/pkg/version.Version=$(VERSION)' \
          -X 'sing-box-web/pkg/version.BuildTime=$(BUILD_TIME)' \
          -X 'sing-box-web/pkg/version.GitCommit=$(GIT_COMMIT)' \
          -X 'sing-box-web/pkg/version.GitBranch=$(GIT_BRANCH)'

# SkyWalking Go Agent配置
SKYWALKING_AGENT_PATH := $(shell which go-agent)
export SW_AGENT_NAME ?= sing-box-web
export SW_AGENT_COLLECTOR_BACKEND_SERVICES ?= skywalking-oap:11800
export SW_AGENT_SAMPLE_N_PER_3_SECS ?= -1

# 默认目标
all: generate build

# 生成代码
generate:
	@echo "Generating code..."
	@go generate ./...
	@$(MAKE) generate-proto
	@$(MAKE) generate-mocks

generate-proto:
	@echo "Generating protobuf code..."
	@mkdir -p api/generated
	@if command -v buf >/dev/null 2>&1; then \
		buf generate; \
	else \
		echo "Warning: buf not found, skipping proto generation"; \
	fi

generate-mocks:
	@echo "Generating mocks..."
	@if command -v mockgen >/dev/null 2>&1; then \
		mockgen -source=internal/api/service/interfaces.go -destination=internal/api/service/mocks/service.go -package=mocks; \
		mockgen -source=internal/api/repository/interfaces.go -destination=internal/api/repository/mocks/repository.go -package=mocks; \
	else \
		echo "Warning: mockgen not found, skipping mock generation"; \
	fi

# 构建应用（默认使用SkyWalking）
build: build-api build-web build-agent

build-api:
	@echo "Building sing-box-api with SkyWalking agent..."
	@mkdir -p $(BUILD_DIR)
	@if [ -n "$(SKYWALKING_AGENT_PATH)" ]; then \
		SW_AGENT_NAME=sing-box-api CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
			-toolexec="$(SKYWALKING_AGENT_PATH)" \
			-ldflags "$(LDFLAGS)" \
			-o $(BUILD_DIR)/sing-box-api \
			./cmd/sing-box-api; \
	else \
		echo "Warning: SkyWalking go-agent not found, building without tracing"; \
		CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
			-ldflags "$(LDFLAGS)" \
			-o $(BUILD_DIR)/sing-box-api \
			./cmd/sing-box-api; \
	fi

build-web:
	@echo "Building sing-box-web with SkyWalking agent..."
	@mkdir -p $(BUILD_DIR)
	@if [ -n "$(SKYWALKING_AGENT_PATH)" ]; then \
		SW_AGENT_NAME=sing-box-web CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
			-toolexec="$(SKYWALKING_AGENT_PATH)" \
			-ldflags "$(LDFLAGS)" \
			-o $(BUILD_DIR)/sing-box-web \
			./cmd/sing-box-web; \
	else \
		echo "Warning: SkyWalking go-agent not found, building without tracing"; \
		CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
			-ldflags "$(LDFLAGS)" \
			-o $(BUILD_DIR)/sing-box-web \
			./cmd/sing-box-web; \
	fi

build-agent:
	@echo "Building sing-box-agent with SkyWalking agent..."
	@mkdir -p $(BUILD_DIR)
	@if [ -n "$(SKYWALKING_AGENT_PATH)" ]; then \
		SW_AGENT_NAME=sing-box-agent CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
			-toolexec="$(SKYWALKING_AGENT_PATH)" \
			-ldflags "$(LDFLAGS)" \
			-o $(BUILD_DIR)/sing-box-agent \
			./cmd/sing-box-agent; \
	else \
		echo "Warning: SkyWalking go-agent not found, building without tracing"; \
		CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
			-ldflags "$(LDFLAGS)" \
			-o $(BUILD_DIR)/sing-box-agent \
			./cmd/sing-box-agent; \
	fi

# 不使用SkyWalking的构建（开发/调试用）
build-no-tracing:
	@echo "Building without SkyWalking agent..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/sing-box-api ./cmd/sing-box-api
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/sing-box-web ./cmd/sing-box-web
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/sing-box-agent ./cmd/sing-box-agent

# 运行测试
test:
	@echo "Running tests..."
	@go test -race -coverprofile=coverage.out ./...

test-integration:
	@echo "Running integration tests..."
	@go test -tags=integration -race ./test/...

# 代码检查
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "Warning: golangci-lint not found, please install it"; \
		go vet ./...; \
	fi

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

# 清理
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -f coverage.out

# 安装开发工具
install-tools:
	@echo "Installing development tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/golang/mock/mockgen@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "Checking if buf is installed..."
	@if ! command -v buf >/dev/null 2>&1; then \
		echo "Installing buf..."; \
		go install github.com/bufbuild/buf/cmd/buf@latest; \
	fi
	@$(MAKE) install-skywalking

# 安装SkyWalking Go Agent
install-skywalking:
	@echo "Installing SkyWalking Go Agent..."
	@if ! command -v go-agent >/dev/null 2>&1; then \
		echo "Installing SkyWalking Go Agent from github.com/apache/skywalking-go/tools/go-agent..."; \
		export PATH=$$PATH:$$HOME/go/bin && go install github.com/apache/skywalking-go/tools/go-agent@latest; \
		echo "SkyWalking Go Agent installed successfully"; \
	else \
		echo "SkyWalking Go Agent already installed: $$(which go-agent)"; \
	fi

# 检查SkyWalking Go Agent状态
check-skywalking:
	@echo "Checking SkyWalking Go Agent..."
	@if command -v go-agent >/dev/null 2>&1; then \
		echo "✓ SkyWalking Go Agent found: $$(which go-agent)"; \
		echo "✓ Version: $$(go-agent --version 2>/dev/null || echo 'Unable to get version')"; \
	else \
		echo "✗ SkyWalking Go Agent not found"; \
		echo "  Run 'make install-skywalking' to install it"; \
		exit 1; \
	fi

# Docker 构建
docker-build: docker-build-api docker-build-web docker-build-agent

docker-build-api:
	@echo "Building sing-box-api Docker image..."
	@docker build -f deployments/docker/api.Dockerfile -t sing-box-api:$(VERSION) .

docker-build-web:
	@echo "Building sing-box-web Docker image..."
	@docker build -f deployments/docker/web.Dockerfile -t sing-box-web:$(VERSION) .

docker-build-agent:
	@echo "Building sing-box-agent Docker image..."
	@docker build -f deployments/docker/agent.Dockerfile -t sing-box-agent:$(VERSION) .

# 发布包
release: clean
	@echo "Building release packages..."
	@mkdir -p $(DIST_DIR)
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			echo "Building $$os/$$arch..."; \
			mkdir -p $(BUILD_DIR)/$$os-$$arch; \
			GOOS=$$os GOARCH=$$arch $(MAKE) build BUILD_DIR=$(BUILD_DIR)/$$os-$$arch; \
			if [ "$$os" = "windows" ]; then \
				mv $(BUILD_DIR)/$$os-$$arch/sing-box-api $(BUILD_DIR)/$$os-$$arch/sing-box-api.exe; \
				mv $(BUILD_DIR)/$$os-$$arch/sing-box-web $(BUILD_DIR)/$$os-$$arch/sing-box-web.exe; \
				mv $(BUILD_DIR)/$$os-$$arch/sing-box-agent $(BUILD_DIR)/$$os-$$arch/sing-box-agent.exe; \
				mv $(BUILD_DIR)/$$os-$$arch/sing-box-ctl $(BUILD_DIR)/$$os-$$arch/sing-box-ctl.exe; \
			fi; \
			tar -czf $(DIST_DIR)/sing-box-web-$(VERSION)-$$os-$$arch.tar.gz -C $(BUILD_DIR)/$$os-$$arch .; \
		done; \
	done

# 运行开发环境
dev-up:
	@echo "Starting development environment..."
	@if [ -f deployments/docker-compose.dev.yml ]; then \
		docker-compose -f deployments/docker-compose.dev.yml up -d; \
	else \
		echo "Development docker-compose file not found"; \
	fi

dev-down:
	@echo "Stopping development environment..."
	@if [ -f deployments/docker-compose.dev.yml ]; then \
		docker-compose -f deployments/docker-compose.dev.yml down; \
	fi

# 数据库迁移
migrate-up:
	@echo "Running database migrations..."
	@if [ -f $(BUILD_DIR)/sing-box-api ]; then \
		./$(BUILD_DIR)/sing-box-api migrate up; \
	else \
		echo "Please build sing-box-api first: make build-api"; \
	fi

migrate-down:
	@echo "Rolling back database migrations..."
	@if [ -f $(BUILD_DIR)/sing-box-api ]; then \
		./$(BUILD_DIR)/sing-box-api migrate down; \
	else \
		echo "Please build sing-box-api first: make build-api"; \
	fi

# 验证生成的代码是否最新
verify-generate: generate
	@echo "Verifying generated code is up-to-date..."
	@if [ -d api/generated ]; then \
		git diff --exit-code api/generated/ || (echo "Generated code is out of date. Please run 'make generate'" && exit 1); \
	fi

# CI 目标
ci: install-tools generate verify-generate lint test

# 运行应用
run-api:
	@echo "Running sing-box-api..."
	@go run ./cmd/sing-box-api serve

run-web:
	@echo "Running sing-box-web..."
	@go run ./cmd/sing-box-web serve

run-agent:
	@echo "Running sing-box-agent..."
	@go run ./cmd/sing-box-agent

# 依赖管理
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  all              - Generate code and build all applications"
	@echo "  generate         - Generate all code (protobuf, mocks, etc.)"
	@echo "  build            - Build all applications with SkyWalking agent"
	@echo "  build-no-tracing - Build all applications without SkyWalking agent"
	@echo "  test             - Run all tests"
	@echo "  lint             - Run code linters"
	@echo "  fmt              - Format code"
	@echo "  clean            - Clean build artifacts"
	@echo "  install-tools    - Install development tools"
	@echo "  install-skywalking - Install SkyWalking Go Agent"
	@echo "  check-skywalking - Check SkyWalking Go Agent installation"
	@echo "  docker-build     - Build Docker images"
	@echo "  release          - Build release packages"
	@echo "  dev-up           - Start development environment"
	@echo "  dev-down         - Stop development environment"
	@echo "  migrate-up       - Run database migrations"
	@echo "  migrate-down     - Rollback database migrations"
	@echo "  ci               - Run CI checks"
	@echo "  run-api          - Run API server in development mode"
	@echo "  run-web          - Run Web server in development mode"
	@echo "  run-agent        - Run Agent in development mode"
	@echo "  deps             - Download and tidy dependencies"
	@echo "  help             - Show this help message"
