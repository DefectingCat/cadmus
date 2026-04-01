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

.PHONY: build build/frontend build/backend build/editor version

# 默认目标：构建全部
build:
	@make build/frontend build/backend

# 显示版本信息
version:
	@echo "App: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Git Branch: $(GIT_BRANCH)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "Platform: $(BUILD_PLATFORM)"

# Build backend Go server
build/backend:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/server ./cmd/server

# Build frontend assets
build/frontend:
	cd web/frontend && bun esbuild src/main.ts \
		--bundle \
		--outdir=../static/dist \
		--minify \
		--format=esm
	cd web/frontend && bun esbuild src/admin/main.ts \
		--bundle \
		--outdir=../static/dist/admin \
		--minify \
		--format=esm
	cd web/frontend && bunx @tailwindcss/cli \
		-i src/styles/main.css \
		-o ../static/dist/styles.css \
		--minify
	cd web/frontend && bunx @tailwindcss/cli \
		-i src/styles/admin.css \
		-o ../static/dist/admin.css \
		--minify

# Generate templ files
build/templ:
	templ generate

# Editor entry point (separate bundle for editor pages)
build/editor:
	cd web/frontend && bun esbuild src/editor/index.ts \
		--bundle \
		--outdir=../static/dist/editor.js \
		--minify

