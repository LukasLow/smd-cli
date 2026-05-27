# Changelog

## [0.2.0] - 2026-05-27

### Added
- **Session mode** (new default): `smd <cmd>` creates a persistent container (`<project>_s`).
  State (apt, pip, npm, files) survives across commands via `docker exec`.
- **Temp mode** (`smd -t <cmd>`): Ephemeral `--rm` container (`<project>_t`), replaces previous default.
- **`smd --close`**: Destroys the session container.
- **New info page**: Designed for AI agents — shows all usage modes at a glance.
- **`SessionConfig`** with `enabled` and `ports` in `smd.toml`.

### Changed
- `smd <cmd>` now uses session mode instead of temp mode by default.
- Temp container gets a name (`<project>_t`) for consistency.
- Live container uses `containerName()` helper for consistent naming.
- `AGENTS.md` now covers ALL commands (ls, cat, apt, git, etc.) through smd.
- README completely rewritten for the new session-mode architecture.

## [0.1.0] - 2026-05-27

### Added
- Persistent package cache volumes (`smd_pkg_npm`, `smd_pkg_pip`, `smd_pkg_go`, `smd_pkg_cargo`, and more)
- New `[volume]` config section in `smd.toml` with per-package toggles
- `buildVolumeArgs()` helper to inject named Docker volumes for package caches
- Volumes are shared across all projects using the same volume name — saves download time
- Automatic volume defaults when initializing from templates

### Changed
- Version bumped to 0.1.0

## [0.0.6] - 2026-05-25

### Added
- `--agentmd` creates/updates AGENTS.md with smd usage instructions
- `runtime` option in `smd.toml` for configurable container runtime

### Changed
- Version bumped to 0.0.6

## [0.0.5] - 2026-05-25

### Changed
- `smd` (no args) now shows an info page
- `smd -o` / `smd --open` starts live mode

### Fixed
- Various bugfixes: nix shell command, env file handling, config saving
