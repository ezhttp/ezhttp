#!/usr/bin/env bash

# Build for all platforms
podman build --platform linux/amd64    --build-arg ARCH=amd64   -t ezhttp/ezhttp-linux-amd64:0.0.3   .;
podman build --platform linux/arm64/v8 --build-arg ARCH=arm64v8 -t ezhttp/ezhttp-linux-arm64v8:0.0.3 .;

# Create Manifest
podman manifest create \
  ezhttp/ezhttp:0.0.3 \
  ezhttp/ezhttp-linux-amd64:0.0.3 \
  ezhttp/ezhttp-linux-arm64v8:0.0.3;

# Publish
podman manifest push ezhttp/ezhttp:0.0.3;
