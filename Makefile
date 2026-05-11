# Salada Makefile

.PHONY: build clean build-arm64

# Default target architecture (change if your server is different)
TARGET_OS ?= linux
TARGET_ARCH ?= amd64

build:
	@echo "Building for $(TARGET_OS)/$(TARGET_ARCH)..."
	@mkdir -p dist
	./scripts/gen-cert.sh
	npm ci
	npm run build
	$(eval GIT_URL := $(shell git config --get remote.origin.url))
	$(eval USERNAME := $(shell echo $(GIT_URL) | sed -E 's/.*[:/]([^/]+)\/[^/]+(\.git)?$$/\1/'))
	$(eval REPO := $(shell echo $(GIT_URL) | sed -E 's/\.git$$//' | sed -E 's/.*\/([^/]+)$$/\1/'))
	$(eval COMMIT := $(shell git rev-parse HEAD))
	CGO_ENABLED=0 GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -ldflags="-s -w -X 'salada/internal/buildinfo.GitUsername=$(USERNAME)' -X 'salada/internal/buildinfo.GitRepo=$(REPO)' -X 'salada/internal/buildinfo.GitCommit=$(COMMIT)'" -o dist/salada .
	@echo "Build complete: dist/salada"
	@ls -lh dist/salada


# Clean build artifacts
clean:
	rm -rf dist/
	@echo "Cleaned dist/"

# Build for different architectures
build-arm64:
	$(MAKE) build TARGET_ARCH=arm64
