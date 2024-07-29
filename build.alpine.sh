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
podman build -t "ezhttp:temp" -f Dockerfile-build --ignorefile .dockerignore-build \
	--build-arg build_date="${BUILD_DATE}" \
	--build-arg build_git_hash="${BUILD_GIT_HASH}" . && \
podman run --rm \
	-v "${PWD}/bin:/output" \
	-w /usr/src/app \
	"ezhttp:temp" \
	cp "/usr/src/app/ezhttp.alpine.bin" /output/ && \
podman image rm "ezhttp:temp";
