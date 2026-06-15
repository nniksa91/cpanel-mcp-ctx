# CLI reference — cpanel-mcp-ctx

## Global flags

| Flag | Default | Description |
|------|---------|-------------|
| `--cpanel-mcp-dir` | `~/.cpanel-mcp` | Override state directory |
| `--json` | off | Machine-readable output |
| `-q`, `--quiet` | off | Errors only |

## Commands

### `init`

Create directory tree with safe permissions.

```bash
cpanel-mcp-ctx init
cpanel-mcp-ctx init --fix-permissions   # repair modes on existing tree
```

Exit `0` if already initialized.

---

### `list`

```bash
cpanel-mcp-ctx list
```

Example output:

```text
PROFILE        ACTIVE  SERVERS  DEFAULT_SERVER   DESCRIPTION
production     *       2        prod-root        WHM prod
majofi-uapi            1        majofi           UAPI majofico
```

---

### `current`

```bash
cpanel-mcp-ctx current
cpanel-mcp-ctx current --json
```

Example:

```json
{
  "profile": "production",
  "config_path": "/home/user/.cpanel-mcp/config.yaml",
  "resolved_path": "/home/user/.cpanel-mcp/configs/production.yaml",
  "default_server": "prod-root",
  "server_count": 2
}
```

---

### `use <profile>`

Switch active profile.

```bash
cpanel-mcp-ctx use majofi-uapi
```

Prints reminder to restart MCP host (Cursor) if process is long-lived.

Exit codes: `0` success, `1` profile missing, `2` validation failed.

---

### `add <profile>`

```bash
cpanel-mcp-ctx add staging --from ~/Downloads/staging-config.yaml
cpanel-mcp-ctx add production --from -    # stdin
cpanel-mcp-ctx add local --copy production
```

Flags:

| Flag | Description |
|------|-------------|
| `--from path` | Source YAML (required unless `--copy`) |
| `--copy <profile>` | Duplicate existing profile |
| `--description text` | Stored in `state.yaml` |
| `--set-current` | Run `use` after add |
| `--allow-inline-token` | Permit configs with `auth.token` (warns loudly) |

---

### `remove <profile>`

```bash
cpanel-mcp-ctx remove staging
cpanel-mcp-ctx remove staging --force   # allowed even if active
```

Refuses to remove active profile without `--force` (leaves symlink dangling — use `--force` only with follow-up `use`).

---

### `show [profile]`

Redacted summary. Default: active profile.

```bash
cpanel-mcp-ctx show
cpanel-mcp-ctx show production
```

Never displays token values. Shows `token_env` names and server host/type table.

---

### `validate [target]`

```bash
cpanel-mcp-ctx validate              # active profile
cpanel-mcp-ctx validate production
cpanel-mcp-ctx validate all
```

Checks:

- YAML parseable
- Top-level `servers` mapping non-empty
- Each server has `type`, `host`, `auth`
- No `token` + `token_env` together
- File/directory permissions
- Optional: overlap in allow/deny lists (mirror MCP loader)

---

### `env [profile]`

```bash
cpanel-mcp-ctx env production --check
# exit 1 if WHM_TOKEN_PROD unset

cpanel-mcp-ctx env production
# prints:
# export WHM_TOKEN_PROD='...'   # only if already set in environment
```

For loading secrets from a vault, document pattern:

```bash
export WHM_TOKEN_PROD="$(pass show whm/prod)"
cpanel-mcp-ctx use production
```

---

### `cursor sync` (removed from scope)

Host-agnostic design: no Cursor-specific JSON editing. Operators point any MCP client at `~/.cpanel-mcp/config.yaml` (symlink) or set `CPANEL_MCP_CONFIG`.

---

## Inline tokens

`add` and `validate` **hard-fail** if any server uses inline `auth.token`. Only `token_env` is accepted.

| Variable | Used by | Description |
|----------|---------|-------------|
| `CPANEL_MCP_CONFIG` | cpanel-mcp | Set by operator or Cursor; switcher may suggest path |
| `CPANEL_MCP_CTX_DIR` | switcher only | Override default `~/.cpanel-mcp` (alternative to flag) |

The switcher does **not** require API tokens in its own environment except when using `env` re-export helpers.
