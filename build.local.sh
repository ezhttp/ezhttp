#!/usr/bin/env bash

# Check Code
go fmt ./...;
go vet ./...;

# Load Dependencies
go mod tidy;

# Build Variables
BUILD_DATE=$(date '+%Y-%m-%d');
BUILD_GIT_HASH=$(git rev-list -1 HEAD);
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')

# Build
mkdir -p ./bin/;
go build -o "./bin/temp.unix.bin" \
	-ldflags="-s -w -X main.BuildDate=${BUILD_DATE} -X main.BuildGoVersion=${GO_VERSION} -X main.BuildGitHash=${BUILD_GIT_HASH}" \
	./main.go;

# Run
"./bin/temp.unix.bin";
