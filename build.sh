#!/usr/bin/env bash

# Install Dependencies
go mod tidy;

# Format and check code
go fmt ./...;
go vet ./...;

# Create bin directory
# mkdir -p ./bin/;

# Build variables
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ");
BUILD_GIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown");

# Build Docker image
podman build \
    --build-arg build_date="${BUILD_DATE}" \
    --build-arg build_git_hash="${BUILD_GIT_HASH}" \
    -t ezhttp:latest \
    .;
