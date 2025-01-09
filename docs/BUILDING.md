# Building

## Requirements

- Golang v1.23.x [Website](https://go.dev/doc/install)
- Podman [Website](https://podman-desktop.io/downloads)

**NOTE:** If you want to use Docker instead, switch `podman` commands to `docker`

## Local Machine

Use this for testing locally.

```bash
# Ensure Podman/Docker is running
./build.local.sh (macOS/Linux)
./build.local.ps1 (Windows)
# Output is in ./bin/ directory
# But it will run automatically
```

## Production (Podman/Docker)

```bash
# Ensure Podman/Docker is running
./build.alpine.sh (macOS/Linux)
./build.alpine.ps1 (Windows)
# Output is in ./bin/ directory
```
