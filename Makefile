# Makefile - Multi-process parallel development mode
# Per docs/design.md Section 11.2

.PHONY: live live/templ live/server live/esbuild live/tailwind build build/frontend

# Simultaneously start all development processes
live:
	@make -j4 live/templ live/server live/esbuild live/tailwind

# templ watch mode (generate Go code)
live/templ:
	templ generate --watch --proxy="http://localhost:8080" --open-browser=false

# Go server (using air hot reload)
live/server:
	air -c .air.toml

# esbuild bundle TypeScript
live/esbuild:
	cd web/frontend && npx esbuild src/main.ts \
		--bundle \
		--outdir=../static/dist \
		--watch \
		--sourcemap=inline \
		--format=esm

# Tailwind CSS compilation
live/tailwind:
	cd web/frontend && npx @tailwindcss/cli \
		-i src/styles/main.css \
		-o ../static/dist/styles.css \
		--watch

# Production build
build/frontend:
	cd web/frontend && npx esbuild src/main.ts \
		--bundle \
		--outdir=../static/dist \
		--minify \
		--format=esm
	cd web/frontend && npx @tailwindcss/cli \
		-i src/styles/main.css \
		-o ../static/dist/styles.css \
		--minify

# Editor entry point (separate bundle for editor pages)
build/editor:
	cd web/frontend && npx esbuild src/editor/index.ts \
		--bundle \
		--outdir=../static/dist/editor.js \
		--minify

live/editor:
	cd web/frontend && npx esbuild src/editor/index.ts \
		--bundle \
		--outdir=../static/dist/editor.js \
		--watch \
		--sourcemap=inline