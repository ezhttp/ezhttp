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

# Build and Copy
podman build -t "ezhttp:temp" -f Dockerfile-build --ignorefile .dockerignore-build `
    --build-arg build_date="$DateTime" `
    --build-arg build_git_hash="$GitHash" . && `
    podman run --rm `
    -v "./bin:/output" `
    -w /usr/src/app `
    "ezhttp:temp" `
    cp "/usr/src/app/ezhttp.alpine.bin" /output/ && `
    podman image rm "ezhttp:temp";
