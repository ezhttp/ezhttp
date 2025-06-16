#!/usr/bin/env bash
set -euo pipefail

# Generate private key and certificate in one command
openssl req -x509 \
    -newkey rsa:4096 \
    -keyout localhost.key \
    -out localhost.crt \
    -days 3650 \
    -nodes \
    -subj "/CN=localhost";

# Set secure permissions
chmod 600 localhost.key
chmod 644 localhost.crt

echo "Generated localhost.key and localhost.crt"
echo "Usage: TLS_KEY=localhost.key TLS_CERT=localhost.crt ./ezhttp"
