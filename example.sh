#!/usr/bin/env bash

# Builds example container and runs it
podman build -t ezhttp-example:latest . && \
    podman run --rm -it -p 8080:8080 ezhttp-example:latest;
