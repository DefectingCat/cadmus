# Makefile - Build commands

.PHONY: build build/frontend build/backend build/editor

# Build all (frontend + backend)
build:
	@make build/frontend build/backend

# Build backend Go server
build/backend:
	go build -o bin/server ./cmd/server

# Build frontend assets
build/frontend:
	cd web/frontend && bun esbuild src/main.ts \
		--bundle \
		--outdir=../static/dist \
		--minify \
		--format=esm
	cd web/frontend && bun @tailwindcss/cli \
		-i src/styles/main.css \
		-o ../static/dist/styles.css \
		--minify

# Editor entry point (separate bundle for editor pages)
build/editor:
	cd web/frontend && bun esbuild src/editor/index.ts \
		--bundle \
		--outdir=../static/dist/editor.js \
		--minify

