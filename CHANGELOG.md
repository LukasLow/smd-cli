# Changelog

## [0.0.5] - 2026-05-25

### Changed
- `smd` (no args) now shows an info page with version and brief usage
- `smd -o` / `smd --open` starts live mode (the old `smd` behavior)
- Version bumped to 0.0.5

### Fixed
- Dynamic nix mode podman command now correctly uses `nix shell` via the image's entrypoint instead of a broken `-c nix run` chain
- Interactive mode selection (temp/live) is now properly reflected in the output message
- Live mode .env file warning now actually shows when env files are mounted outside the allow list
- Removed dead `nixArgs` block in podman.go that was built but never used
- SaveConfig now uses atomic write (temp file + rename) to prevent config corruption
- Port deduplication in `defaultPorts` to avoid duplicate `8080:8080` entries
- Removed duplicate `defaultBlockFiles` function (consolidated into `defaultBlockFilesForTypes`)
- Fixed variable shadowing in main.go (`var err error` in inner scope)
- Fixed stale comment and improved error messages

### Maintenance
- Added test suite for core functions (`main_test.go`)
- Fixed module path in go.mod (`github.com/user/smd` → `github.com/LukasLow/smd-cli`)
- Fixed homepage URL in flake.nix
- Updated README to reflect `smd --open` for live mode
- Added CHANGELOG.md
