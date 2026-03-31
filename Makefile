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

