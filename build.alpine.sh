#!/usr/bin/env bash

# Delint
go fmt ./...;

# Build Variables
BUILD_DATE=$(date '+%Y-%m-%d');
BUILD_GIT_HASH=$(git rev-list -1 HEAD);

# Issues? Restart Podman Machine
# podman machine stop; podman machine start

# Build Container Image. We'll copy out the file after
mkdir -p ./bin/;
# AMD64 x86_64
podman build --platform linux/amd64 -t "ezhttp:temp-amd64" -f Dockerfile-build --ignorefile .dockerignore-build \
	--build-arg build_date="${BUILD_DATE}" \
	--build-arg build_git_hash="${BUILD_GIT_HASH}" . && \
	podman run --rm \
	-v "${PWD}/bin:/output" \
	-w /usr/src/app \
	"ezhttp:temp-amd64" \
	cp "/usr/src/app/ezhttp.alpine.bin" /output/ezhttp.alpine-amd64.bin && \
	podman image rm "ezhttp:temp-amd64";
# ARM64 v8
podman build --platform linux/arm64/v8 -t "ezhttp:temp-arm64v8" -f Dockerfile-build --ignorefile .dockerignore-build \
	--build-arg build_date="${BUILD_DATE}" \
	--build-arg build_git_hash="${BUILD_GIT_HASH}" . && \
	podman run --rm \
	-v "${PWD}/bin:/output" \
	-w /usr/src/app \
	"ezhttp:temp-arm64v8" \
	cp "/usr/src/app/ezhttp.alpine.bin" /output/ezhttp.alpine-arm64v8.bin && \
	podman image rm "ezhttp:temp-arm64v8";
