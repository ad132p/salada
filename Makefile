# Salada deployment Makefile
# Simple commands for building and deploying

.PHONY: build deploy clean

# Default target architecture (change if your server is different)
TARGET_OS ?= linux
TARGET_ARCH ?= amd64

build:
	@echo "Building for $(TARGET_OS)/$(TARGET_ARCH)..."
	@mkdir -p dist
	pnpm install --frozen-lockfile
	pnpm run build
	CGO_ENABLED=0 GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -ldflags="-s -w" -o dist/salada .
	@echo "Build complete: dist/salada"
	@ls -lh dist/salada

# Usage: make deploy SERVER=user@example.com
deploy: build
	@if [ -z "$(SERVER)" ]; then \
		echo "Error: SERVER not set. Usage: make deploy SERVER=user@example.com"; \
		exit 1; \
	fi
	@echo "Deploying to $(SERVER)..."
	./scripts/deploy.sh $(SERVER)

# Quick deploy without rebuild (use after make build)
# Usage: make quick-deploy SERVER=user@example.com
quick-deploy:
	@if [ -z "$(SERVER)" ]; then \
		echo "Error: SERVER not set. Usage: make quick-deploy SERVER=user@example.com"; \
		exit 1; \
	fi
	./scripts/quick-deploy.sh $(SERVER)

# Clean build artifacts
clean:
	rm -rf dist/
	@echo "Cleaned dist/"

# Build for different architectures
build-arm64:
	$(MAKE) build TARGET_ARCH=arm64
