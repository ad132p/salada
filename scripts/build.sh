#!/bin/bash
# Build script for salada - run this locally, NOT on the server
# Cross-compiles the Go binary and builds frontend assets

set -e

echo "=== Salada Build Script ==="
echo "Building locally for remote deployment..."

# Get target architecture (default: linux/amd64 for most servers)
TARGET_OS=${TARGET_OS:-linux}
TARGET_ARCH=${TARGET_ARCH:-amd64}

echo "Target: $TARGET_OS/$TARGET_ARCH"

# Build frontend first (needs to happen before Go embed)
echo "Building frontend..."
./scripts/gen-cert.sh
npm install --frozen-lockfile
npm run build

# Build Go binary with cross-compilation
echo "Building Go binary..."
CGO_ENABLED=0 GOOS=$TARGET_OS GOARCH=$TARGET_ARCH go build -ldflags="-s -w" -o dist/salada .

echo "Build complete! Artifacts in ./dist/"
ls -lh dist/
