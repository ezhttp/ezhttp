#!/usr/bin/env bash

# Generate private key and certificate in one command
openssl req -x509 \
    -newkey rsa:2048 \
    -keyout localhost.key \
    -out localhost.crt \
    -days 3650 \
    -nodes \
    -subj "/CN=localhost";

echo "Generated localhost.key and localhost.crt";
echo "Run with TLS_KEY=localhost.key TLS_CERT=localhost.crt go run ./cmd/server/main.go";
