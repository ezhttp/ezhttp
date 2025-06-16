#!/usr/bin/env bash

# Publish script for multi-arch Docker images
# Usage: ./publish.sh [version]
# Example: ./publish.sh 0.0.4

VERSION=${1:-latest};

# Build variables
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ");
BUILD_GIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown");

# Build for multiple platforms
echo "Building multi-arch images for version: ${VERSION}";

# AMD64
podman build --platform linux/amd64 \
    --build-arg build_date="${BUILD_DATE}" \
    --build-arg build_git_hash="${BUILD_GIT_HASH}" \
    -t "ezhttp/ezhttp-linux-amd64:${VERSION}" .;

# ARM64
podman build --platform linux/arm64/v8 \
    --build-arg build_date="${BUILD_DATE}" \
    --build-arg build_git_hash="${BUILD_GIT_HASH}" \
    -t "ezhttp/ezhttp-linux-arm64v8:${VERSION}" .;

# Create manifest
podman manifest create \
    "ezhttp/ezhttp:${VERSION}" \
    "ezhttp/ezhttp-linux-amd64:${VERSION}" \
    "ezhttp/ezhttp-linux-arm64v8:${VERSION}";

# Push manifest
echo "Pushing manifest...";
podman manifest push "ezhttp/ezhttp:${VERSION}";

echo "Published ezhttp:${VERSION}";
