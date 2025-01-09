<#
.SYNOPSIS
    Builds ezhttp with Podman/Alpine
#>

# Variables
$DateTime = (Get-Date).ToString("yyyyMMdd")
$GitHash = (git log -1 --format=format:"%H")
# Can also truncate
#$GitHash = $GitHashFull.Substring(0, 8)

# Create ./bin/
New-Item -ItemType Directory -Force -Path ./bin/ | out-null;

# Build Container Image. We'll copy out the file after
# AMD64 x86_64
podman build --platform linux/amd64 -t "ezhttp:temp-amd64" -f Dockerfile-build --ignorefile .dockerignore-build `
    --build-arg build_date="$DateTime" `
    --build-arg build_git_hash="$GitHash" . && `
    podman run --rm `
    -v "./bin:/output" `
    -w /usr/src/app `
    "ezhttp:temp-amd64" `
    cp "/usr/src/app/ezhttp.alpine.bin" /output/ezhttp.alpine-amd64.bin && `
    podman image rm "ezhttp:temp-amd64";
# ARM64 v8
podman build --platform linux/arm64/v8 -t "ezhttp:temp-arm64v8" -f Dockerfile-build --ignorefile .dockerignore-build `
    --build-arg build_date="$DateTime" `
    --build-arg build_git_hash="$GitHash" . && `
    podman run --rm `
    -v "./bin:/output" `
    -w /usr/src/app `
    "ezhttp:temp-arm64v8" `
    cp "/usr/src/app/ezhttp.alpine.bin" /output/ezhttp.alpine-arm64v8.bin && `
    podman image rm "ezhttp:temp-arm64v8";
