# Makefile - Build commands

# 版本信息
APP_NAME := cadmus
VERSION := 0.0.1
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
GO_VERSION := $(shell go version | awk '{print $$3}')
BUILD_PLATFORM := $(shell go env GOOS)/$(shell go env GOARCH)
BUILD_DIR := bin

# 生产构建标志
LDFLAGS := -s -w \
    -X 'main.version=$(VERSION)' \
    -X 'main.gitCommit=$(GIT_COMMIT)' \
    -X 'main.gitBranch=$(GIT_BRANCH)' \
    -X 'main.buildTime=$(BUILD_TIME)' \
    -X 'main.goVersion=$(GO_VERSION)' \
    -X 'main.buildPlatform=$(BUILD_PLATFORM)'

.PHONY: build build/frontend build/backend build/editor version test test/coverage test/bench

# 默认目标：构建全部
build:
	@make build/frontend build/backend

# Build backend Go server
build/backend:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/server ./cmd/server

# Build frontend assets using package.json scripts
build/frontend:
	cd web/frontend && bun run build:all

# Generate templ files
build/templ:
	templ generate

# Editor entry point (separate bundle for editor pages)
build/editor:
	cd web/frontend && bun esbuild src/editor/index.ts \
		--bundle \
		--outdir=../static/dist/editor.js \
		--minify

# 运行所有测试（带竞态检测）
test:
	go test -v -race ./...

# 生成覆盖率报告
test/coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# 运行基准测试
test/bench:
	go test -bench=. -benchmem ./...

# 格式化代码（使用 goimports 替代 go fmt）
fmt:
	@echo "Formatting code with goimports..."
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w $(shell find . -path './data' -prune -o -name '*.go' -type f -print); \
	else \
		echo "goimports not installed. Run: go install golang.org/x/tools/cmd/goimports@latest"; \
		exit 1; \
	fi

# 静态检查
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./cmd/... ./internal/... ./pkg/... ./plugins/... ./test/...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./cmd/... ./internal/... ./pkg/... ./plugins/... ./test/...; \
	fi

# 显示版本信息
version:
	@echo "App: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Git Branch: $(GIT_BRANCH)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "Platform: $(BUILD_PLATFORM)"
