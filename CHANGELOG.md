# Changelog

All notable changes to **cpanel-mcp-ctx** are documented in this file.

Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Changed

- Publish and distribute via **GitHub** (`github.com/nniksa91/cpanel-mcp-ctx`) instead of
  private GitLab; Go module path updated accordingly.
- CI: GitHub Actions (test + GoReleaser release on tag).

## [0.1.0] - 2026-06-13

### Added

- Go CLI `cpanel-mcp-ctx` with profile management under `~/.cpanel-mcp/configs/`.
- Commands: `init`, `list`, `current`, `use`, `add`, `remove`, `show`, `validate`, `env`.
- Active profile switching via atomic symlink (`config.yaml` → `configs/<profile>.yaml`).
- YAML validation aligned with cPanel MCP config schema; hard-fail on inline `auth.token`.
- `--json` and `--quiet` global flags; host-agnostic operator documentation.
- CI workflow (test, vet, golangci-lint, govulncheck) and GoReleaser config.
- Project scaffold and `documentation/` design pack.
