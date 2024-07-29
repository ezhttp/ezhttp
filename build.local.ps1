<#
.SYNOPSIS
    Used ONLY for local testing on Windows

.DESCRIPTION
    Use this ONLY to build and test on Windows machines.
    All production builds should be made in a container using an Alpine base image
#>

# Check Code
go fmt ./...;
go vet ./...;

# Load Dependencies
go mod tidy;

# Build
New-Item -ItemType Directory -Force -Path .\bin\ | out-null;
go build -o ".\bin\temp.windows.exe" `
    -ldflags="-s -w -X main.BuildDate=WINDOWS -X main.BuildGoVersion=WINDOWS -X main.BuildGitHash=WINDOWS" `
    ./main.go;

# Run
.\bin\temp.windows.exe;
