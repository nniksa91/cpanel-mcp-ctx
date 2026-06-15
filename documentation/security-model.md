# Security model — cpanel-mcp-ctx

This tool manages **paths to credentials**, not credentials themselves. Security goals align with cPanel MCP’s operator checklist (`security-audit.md`).

---

## Threat model

### In scope

| Threat | Mitigation |
|--------|------------|
| World-readable config files | Enforce `0600` on YAML, `0700` on dirs; `validate` checks modes |
| Token leakage via CLI output | `show` uses redaction; never print `auth.token` values |
| Symlink hijacking | Resolve paths under `--cpanel-mcp-dir`; refuse if dir is symlink to untrusted location (configurable strict mode) |
| TOCTOU on switch | Atomic rename for symlink and state file |
| Accidental commit of profiles | Document `.gitignore`; CLI never writes outside cpanel-mcp dir |
| Profile name injection | Strict charset validation on profile names |

### Out of scope (v1)

- Compromised MCP host / LLM prompt injection
- Network MITM (handled by `verify_tls` in MCP server)
- Stolen env vars from process memory

---

## Secrets handling rules (implementation MUST)

1. **Prefer `token_env`** — on `add`, warn if `auth.token` present; optional `--allow-inline-token` for advanced users.
2. **Never log** environment variable values, inline tokens, or full YAML dumps at info level.
3. **`env` subcommand** — print only variable **names** in check mode; value mode only re-exports already-set vars (same as `printenv` guard).
4. **No cloud upload** — no telemetry, no phone-home.
5. **No shell eval** — output `export FOO=...` as text; document that user should `eval "$(cpanel-mcp-ctx env production)"` consciously.

---

## File permissions

Created by `init` and enforced on `add`:

```text
drwx------  ~/.cpanel-mcp/
drwx------  ~/.cpanel-mcp/configs/
-rw-------  ~/.cpanel-mcp/configs/*.yaml
-rw-------  ~/.cpanel-mcp/state.yaml
lrwx------  ~/.cpanel-mcp/config.yaml  → configs/<profile>.yaml
```

On Linux, umask should be `077` during writes. If existing files are too permissive, `validate` reports and `init --fix-permissions` repairs.

---

## Cursor `mcp.json` sync (v0.2 — optional)

If implemented:

- Backup to `mcp.json.bak.<timestamp>` before write
- Only touch `env.CPANEL_MCP_CONFIG` and never token values
- Require `--yes` or interactive confirm
- Validate JSON schema before save

---

## Supply chain

- Go module checksums in `go.sum`
- CI: `govulncheck`, `golangci-lint`
- Release: signed tags + GitHub checksums file (GoReleaser)
- Minimal dependencies: `cobra`, `yaml.v3` — audit regularly

---

## Compliance mapping (operator controls)

| Control | cpanel-mcp-ctx contribution |
|---------|----------------------------|
| Least privilege | Encourages separate profiles per token scope |
| Secrets not at rest in config | Validates/warns on inline tokens |
| Config file protection | 0600 enforcement |
| Audit trail | Optional: append-only log of profile switches (v0.3) |

Not a SOC 2 certification — aids operator hygiene only.

---

## Security review checklist (pre-release)

- [ ] Grep codebase for `fmt.Print` of config/auth structs
- [ ] Fuzz profile names for path traversal (`../etc/passwd`)
- [ ] Test symlink points outside cpanel-mcp dir → rejected
- [ ] Test world-readable config → `validate` fails with clear message
- [ ] Run with `strace` / logging — confirm no unexpected file reads

Use **Security Engineer** agent (`.cursor/rules/security-engineer.mdc`) for final pass.
