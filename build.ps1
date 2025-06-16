<#
.SYNOPSIS
    Build script for ezhttp Docker image
#>

# Install Dependencies
go mod tidy;

# Format and check code
go fmt ./...;
go vet ./...;

# Create bin directory
# New-Item -ItemType Directory -Force -Path ./bin/ | Out-Null;

# Build variables
# Can also truncate
#$GitHash = $GitHashFull.Substring(0, 8)
$BUILD_DATE = Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ" -AsUTC
$BUILD_GIT_HASH = git rev-parse --short HEAD 2>$null
if (-not $BUILD_GIT_HASH) { $BUILD_GIT_HASH = "unknown" }

# Build Docker image
podman build `
    --build-arg build_date="$BUILD_DATE" `
    --build-arg build_git_hash="$BUILD_GIT_HASH" `
    -t ezhttp:latest `
    .
