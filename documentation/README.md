# Documentation index — cpanel-mcp-ctx

| Document | Description |
|----------|-------------|
| [architecture.md](architecture.md) | On-disk layout, symlink switching, MCP integration |
| [security-model.md](security-model.md) | Threat model, permissions, secrets policy |
| [cli-reference.md](cli-reference.md) | Command reference |
| [distribution.md](distribution.md) | GitLab packaging, CI, GoReleaser, install paths |

User-facing guides (onboarding, AI clients, GitLab install): [`../../Readmes/`](../../Readmes/)

## Quick summary

- **Language:** Go 1.22+ (`cpanel-mcp-ctx` single binary)
- **Model:** profiles as separate YAML files + active symlink
- **Secrets:** `token_env` only; switcher never stores tokens
- **CI:** `.github/workflows/` on [github.com/nniksa91/cpanel-mcp-ctx](https://github.com/nniksa91/cpanel-mcp-ctx)
