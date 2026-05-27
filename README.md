# smd - Containerized Development for AI Agents

A CLI tool that wraps Docker/Podman to provide secure, persistent containerized environments with automatic protection of sensitive `.env` files. Designed for AI agents that run multiple commands.

[![Release](https://img.shields.io/github/v/release/LukasLow/smd-cli)](https://github.com/LukasLow/smd-cli/releases)

## Installation

### macOS / Linux (Binary Download)
```bash
curl -sSL https://raw.githubusercontent.com/LukasLow/smd-cli/main/install.sh | bash
```

### Using Nix
```bash
nix profile install github:LukasLow/smd-cli
```

### Build from Source
```bash
git clone https://github.com/LukasLow/smd-cli.git
cd smd-cli
go build -o smd ./src/
```

## Quick Start

```bash
# Just run any command — auto-detects the environment
smd npm install        # Persistent session, npm deps cached
smd python script.py   # Python in persistent environment
smd apt install curl   # System packages survive across commands

# Ephemeral (fresh container each time)
smd -t python test.py

# Interactive development
smd --open

# Destroy session container
smd --close
```

## How It Works

### Session Mode (default)
```
smd <command> [args...]
```
Creates a persistent named container (`<project>_s`). All commands run in the same container via `docker exec`. State survives — apt packages, pip installs, npm dependencies, file changes.

- First run: container is created automatically
- Subsequent runs: container is started, command executed, container stopped
- Use `smd --close` to destroy the session

### Temp Mode
```
smd -t <command> [args...]
```
Runs a command in an ephemeral container (`<project>_t`, `--rm`). Fresh container every time. Use when isolation matters.

### Live Mode
```
smd --open
```
Interactive shell with PWD mounted read-write. Ideal for long-running dev servers and file watching.

## Package Cache Volumes

Named volumes persist package downloads across projects:

| Volume | Cache Path | Used By |
|--------|-----------|---------|
| `smd_pkg_npm` | `/root/.npm` | npm, yarn, pnpm |
| `smd_pkg_yarn` | `/root/.yarn` | Yarn |
| `smd_pkg_pnpm` | `/root/.local/share/pnpm` | pnpm |
| `smd_pkg_pip` | `/root/.cache/pip` | pip |
| `smd_pkg_poetry` | `/root/.cache/pypoetry` | Poetry |
| `smd_pkg_go` | `/go/pkg/mod` | Go modules |
| `smd_pkg_cargo` | `/usr/local/cargo/registry` | Cargo |
| `smd_pkg_bun` | `/root/.bun/install/cache` | Bun |
| `smd_pkg_conda` | `/opt/conda/pkgs` | Conda |

### Safety Shield
- Automatically detects `.env` files in your project
- In **temp mode**: All `.env` files are blocked unless whitelisted
- In **session/live mode**: Files are available with a warning

## Templates

```bash
smd --init
```

| Template | Description |
|----------|-------------|
| `full` | Everything installed (all languages) |
| `dynamic` | NixOS base — install any tool via `nix shell` |
| `javascript/*` | Node, Bun, all, demo |
| `python/*` | Python, pip, conda, all |
| `go`, `go/air` | Go standard, Go with Air |
| `rust` | Rust with cargo |
| `elixir`, `gleam` | Other languages |

## Configuration

Minimal `smd.toml` — most fields have smart defaults:

```toml
[project]
type = ["nodejs"]

[security]
allow_env = [".env.local"]

[volume.packages]
npm = true
```

Or leave it out entirely — smd auto-detects from the command.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `SMD_IMAGE` | Override the container image |
| `SMD_DEBUG` | Show podman/docker commands being executed |
| `SMD_DYNAMIC_PACKAGE` | Package for dynamic nix shell mode |

## Requirements

- Docker or Podman
- Go 1.21+ (for building)
