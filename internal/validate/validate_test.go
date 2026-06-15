package validate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nniksa91/cpanel-mcp-ctx/internal/validate"
)

func TestValidateValidFixture(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "valid.yaml")
	summary, err := validate.ValidateFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if summary.DefaultServer != "prod-root" {
		t.Fatalf("default_server = %q", summary.DefaultServer)
	}
	if len(summary.Servers) != 2 {
		t.Fatalf("servers = %d", len(summary.Servers))
	}
	envs := validate.TokenEnvs(summary)
	if len(envs) != 2 {
		t.Fatalf("token envs = %v", envs)
	}
}

func TestRejectInlineToken(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "inline-token.yaml")
	_, err := validate.ValidateFile(path)
	if err == nil {
		t.Fatal("expected error for inline token")
	}
}

func TestRejectBothTokenFields(t *testing.T) {
	raw := []byte(`
servers:
  s:
    type: whm
    host: whm.example.com
    auth:
      method: token
      username: root
      token: x
      token_env: T
`)
	_, err := validate.ValidateBytes(raw, "test")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckPermissions(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	warnings := validate.CheckPermissions(dir)
	if len(warnings) == 0 {
		t.Fatal("expected permission warnings")
	}
}
