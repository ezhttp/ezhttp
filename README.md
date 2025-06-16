# ezhttp

## Purpose

Lightweight webserver for containers in Golang with security and optimizations

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

- [Building](docs/BUILDING.md)
- [Resources](docs/RESOURCES.md)
