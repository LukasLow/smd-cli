# smd - Secure My Directory

A CLI tool that wraps Podman to provide secure, containerized development environments with automatic protection of sensitive `.env` files.

[![Release](https://img.shields.io/github/v/release/LukasLow/smd-cli)](https://github.com/LukasLow/smd-cli/releases)

## Installation

### macOS / Linux (Binary Download)
```bash
# Download latest release (adjust OS and ARCH as needed)
curl -LO https://github.com/LukasLow/smd-cli/releases/latest/download/smd-linux-amd64
chmod +x smd-linux-amd64
sudo mv smd-linux-amd64 /usr/local/bin/smd

# Or use the install script
curl -sSL https://raw.githubusercontent.com/LukasLow/smd-cli/main/install.sh | bash
```

### Using Nix (recommended for Nix users)
```bash
# Install via flake
nix profile install github:LukasLow/smd-cli

# Or clone and build
nix build github:LukasLow/smd-cli
./result/bin/smd --version
```

### Build from Source
```bash
git clone https://github.com/LukasLow/smd-cli.git
cd smd-cli
go build -o smd ./src/
```

## Quick Start

```bash
# Initialize with template selection
smd --init

# Or just run a command - auto-detects the environment
smd npm install     # Auto-detects Node.js
smd python script.py # Auto-detects Python
smd go run .        # Auto-detects Go or uses dynamic mode

# Live development mode (interactive shell)
smd
```

## Usage

### Live Mode
```bash
smd
```
Starts an interactive container with your PWD mounted at `/app`. Ideal for development with file watching and long-running processes.

### Temp/Dynamic Mode
```bash
smd <command> [args...]
```
Runs a command in an ephemeral container. If no `smd.toml` exists:
- **Auto-detect**: Recognizes commands like `npm`, `python`, `go` and uses appropriate template
- **Dynamic mode**: Unknown commands use NixOS + `nix shell` to install on-demand

Sensitive files (`.env`, `.npmrc`, etc.) are automatically blocked in temp mode.

## Templates

Initialize a project with `smd --init` and select a template:

| Template | Description |
|----------|-------------|
| `full` | Everything installed (all languages) |
| `dynamic` | NixOS base - install any tool via `nix shell` |
| `javascript/all` | Node + Bun + Deno |
| `javascript/node` | Node.js only |
| `javascript/bun` | Bun runtime |
| `python` | Python with pip |
| `python/all` | Python + poetry + conda |
| `go` | Go standard |
| `go/air` | Go with Air live-reload |
| `rust` | Rust with cargo |
| `elixir` | Elixir/Erlang |
| `gleam` | Gleam language |

## How It Works

### Safety Shield
- Automatically detects `.env` files in your project
- In **temp mode**: All `.env` files are blocked unless explicitly whitelisted in `smd.toml`
- In **live mode**: Files are available but a warning is shown

### Modes

| Mode | Use Case | PWD Mount | .env Files |
|------|----------|-----------|------------|
| Live | Development, dev servers | Yes (read-write) | Available with warning |
| Temp | Install, build, one-off | Yes (configurable) | Blocked by default |

## Configuration (smd.toml)

```toml
[project]
name = "my-app"
type = ["nodejs"]

[container]
image = "docker.io/node:20-alpine"
workdir = "/app"

[security]
allow_env = [".env.local"]  # Whitelist for temp mode
block_files = [".npmrc", ".yarnrc"]

[live]
mount_pwd = true
ports = ["3000:3000", "8080:8080"]
```

## Requirements

- Podman
- Go 1.21+ (for building)

## Environment Variables

| Variable | Description |
|----------|-------------|
| `SMD_IMAGE` | Override the container image |
| `SMD_DEBUG` | Show podman commands being executed |
| `SMD_DYNAMIC_PACKAGE` | Package for dynamic nix shell mode |
