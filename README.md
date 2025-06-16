# ezhttp

## Purpose

Lightweight webserver for containers in Golang with security and optimizations

## Requirements

- Golang v1.24.x [Website](https://go.dev/doc/install)
- Podman [Website](https://podman-desktop.io/downloads)

**NOTE:** If you want to use Docker instead, switch `podman` commands to `docker`

## Quick Start

Build and run the example container:

```bash
# Using Podman
podman build -t ezhttp-example:latest . && \
    podman run --rm -it -p 8080:8080 ezhttp-example:latest

# Using Docker
docker build -t ezhttp-example:latest . && \
    docker run --rm -it -p 8080:8080 ezhttp-example:latest
```

Then open http://localhost:8080 in your browser.

## Links

- [Resources](docs/RESOURCES.md)
