#!/usr/bin/env bash
set -euo pipefail

# Build script for EZhttp
echo "Building EZhttp...";

# Clean previous builds
rm -f ezhttp-server ezhttp-proxy;

# Build server
echo "Building server...";
go build -o ezhttp-server ./cmd/server;

# Build proxy
echo "Building proxy...";
go build -o ezhttp-proxy ./cmd/proxy;

echo "Build complete!";
echo "";
echo "Binaries created:";
echo "  - ezhttp-server (web server)";
echo "  - ezhttp-proxy (reverse proxy)";
