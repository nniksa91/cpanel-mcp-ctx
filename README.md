# cpanel-mcp-ctx

CLI profile switcher for [@terynas/cpanel-mcp](https://www.npmjs.com/package/@terynas/cpanel-mcp) — manage multiple `~/.cpanel-mcp/configs/*.yaml` profiles and switch the active target with one command.

Works with **any MCP host** (Cursor, Claude Desktop, VS Code, ChatGPT, etc.) that launches `cpanel-mcp` with `--config` or the default `~/.cpanel-mcp/config.yaml` path. No host-specific integration.

**Repository:** [github.com/nniksa91/cpanel-mcp-ctx](https://github.com/nniksa91/cpanel-mcp-ctx)

## Install

### Release binary (recommended)

Download from [GitHub Releases](https://github.com/nniksa91/cpanel-mcp-ctx/releases):

```bash
VERSION=0.1.0
PLATFORM=linux_amd64   # or linux_arm64, darwin_amd64, darwin_arm64, windows_amd64, windows_arm64

curl -fsSL \
  "https://github.com/nniksa91/cpanel-mcp-ctx/releases/download/v${VERSION}/cpanel-mcp-ctx_${VERSION}_${PLATFORM}.tar.gz" \
  | tar -xz -C ~/.local/bin cpanel-mcp-ctx
chmod +x ~/.local/bin/cpanel-mcp-ctx
cpanel-mcp-ctx --help
```

On Windows, download the `.zip` asset from the release page and add `cpanel-mcp-ctx.exe` to your `PATH`.

### Go install

```bash
go install github.com/nniksa91/cpanel-mcp-ctx/cmd/cpanel-mcp-ctx@latest
```

Requires Go 1.22+.

### Build from source

```bash
git clone https://github.com/nniksa91/cpanel-mcp-ctx.git
cd cpanel-mcp-ctx
make install    # ~/.local/bin/cpanel-mcp-ctx
```

## Quick start

```bash
cpanel-mcp-ctx init
cpanel-mcp-ctx add production --from ~/.cpanel-mcp/config.yaml.backup --description "WHM prod"
cpanel-mcp-ctx add staging --from ./staging.yaml --set-current
cpanel-mcp-ctx list
cpanel-mcp-ctx use production
```

After switching, **restart your MCP host process** so `cpanel-mcp` reloads the config.

### MCP host configuration (generic)

Point your MCP client at the stable symlink path:

```json
{
  "command": "cpanel-mcp",
  "args": ["--config", "/home/USER/.cpanel-mcp/config.yaml"],
  "env": {
    "WHM_TOKEN_PROD": "...",
    "CPANEL_UAPI_TOKEN": "..."
  }
}
```

Keep API tokens in environment variables (`token_env` in YAML). Inline `auth.token` values are **rejected**.

## Commands

| Command | Description |
|---------|-------------|
| `init` | Create `~/.cpanel-mcp/` tree (`0700`/`0600`) |
| `add <profile> --from file` | Install a profile YAML |
| `use <profile>` | Switch active profile (symlink) |
| `list` | List profiles |
| `current` | Show active profile |
| `show [profile]` | Redacted summary (no token values) |
| `validate [profile\|all]` | Validate YAML + permissions |
| `env [profile] --check` | Verify `token_env` vars are set |
| `remove <profile>` | Delete a profile |

Global flags: `--cpanel-mcp-dir`, `--json`, `--quiet`.

See [`documentation/cli-reference.md`](documentation/cli-reference.md) for full details.

## Security

- Config files are stored with mode `0600`, directories with `0700`.
- The CLI never prints API token values.
- Inline `auth.token` in YAML is hard-failed on `add` and `validate`.

See [`documentation/security-model.md`](documentation/security-model.md).

## Documentation

| Document | Topic |
|----------|-------|
| [documentation/architecture.md](documentation/architecture.md) | On-disk layout |
| [documentation/distribution.md](documentation/distribution.md) | Packaging and CI |
| [documentation/cli-reference.md](documentation/cli-reference.md) | All commands |

## License

MIT — see [LICENSE](LICENSE).
