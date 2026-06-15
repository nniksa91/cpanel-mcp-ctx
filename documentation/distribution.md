# Distribution — cpanel-mcp-ctx

Standalone companion to cPanel MCP. Distributed via **GitHub Releases** under [nniksa91/cpanel-mcp-ctx](https://github.com/nniksa91/cpanel-mcp-ctx).

---

## Repository identity

| Item | Value |
|------|-------|
| Folder (monorepo) | `mcp-context-switcher/` |
| Git remote | `https://github.com/nniksa91/cpanel-mcp-ctx` |
| Go module | `github.com/nniksa91/cpanel-mcp-ctx` |
| Binary name | `cpanel-mcp-ctx` |
| License | MIT |

---

## Install methods

### 1. GitHub Releases (operators) — **recommended**

No auth required for public releases.

```bash
VERSION=0.1.0
PLATFORM=linux_amd64

curl -fsSL \
  "https://github.com/nniksa91/cpanel-mcp-ctx/releases/download/v${VERSION}/cpanel-mcp-ctx_${VERSION}_${PLATFORM}.tar.gz" \
  | tar -xz -C ~/.local/bin cpanel-mcp-ctx
chmod +x ~/.local/bin/cpanel-mcp-ctx
```

### 2. Go install (developers)

```bash
go install github.com/nniksa91/cpanel-mcp-ctx/cmd/cpanel-mcp-ctx@latest
```

Requires Go 1.22+.

### 3. Release binaries (CI)

GoReleaser on tag `v*` uploads to GitHub Releases.

| OS | Arch | Artifact |
|----|------|----------|
| linux | amd64, arm64 | `cpanel-mcp-ctx_<ver>_linux_<arch>.tar.gz` |
| darwin | amd64, arm64 | `cpanel-mcp-ctx_<ver>_darwin_<arch>.tar.gz` |
| windows | amd64, arm64 | `cpanel-mcp-ctx_<ver>_windows_<arch>.zip` |

Each archive contains `cpanel-mcp-ctx`, `LICENSE`, `README.md`, and `SHA256SUMS`.

---

## CI pipeline (GitHub Actions)

| Workflow | Trigger | Job |
|----------|---------|-----|
| `.github/workflows/ci.yml` | push / PR | `go test`, `go vet`, build, govulncheck |
| `.github/workflows/release.yml` | tag `v*` | GoReleaser → GitHub Releases |

Tag format: `v0.1.0` — triggers GoReleaser with `GITHUB_TOKEN`.

---

## Versioning

Semantic Versioning:

- **0.1.0** — MVP: init, list, current, use, add, remove, show, validate, env
- **0.2.0** — shell completions
- **1.0.0** — stable CLI contract

---

## Operator quick start (post-install)

```bash
cpanel-mcp-ctx init
cpanel-mcp-ctx add production --from ~/.cpanel-mcp/config.yaml.backup
cpanel-mcp-ctx add majofi-uapi --from ./majofi.yaml --set-current
cpanel-mcp-ctx list
cpanel-mcp-ctx use production
# restart MCP host or reload window
```

Ensure token env vars are set in the MCP client config — switcher does not store tokens.

---

## Relationship to cPanel MCP

| Component | Install separately? |
|-----------|---------------------|
| `cpanel-mcp` (MCP server) | Yes — `npm install -g @terynas/cpanel-mcp` |
| `cpanel-mcp-ctx` (this tool) | Yes — binary from GitHub Releases |
