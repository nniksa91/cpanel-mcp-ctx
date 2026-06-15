package store_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nniksa91/cpanel-mcp-ctx/internal/store"
)

func openTestStore(t *testing.T) *store.Store {
	t.Helper()
	base := t.TempDir()
	st, err := store.Open(base)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Init(false); err != nil {
		t.Fatal(err)
	}
	return st
}

func readFixture(t *testing.T) *os.File {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "testdata", "valid.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func TestInitAddUseCurrent(t *testing.T) {
	st := openTestStore(t)
	if err := st.Add("production", "prod", readFixture(t), false); err != nil {
		t.Fatal(err)
	}
	if err := st.Use("production"); err != nil {
		t.Fatal(err)
	}
	cur, err := st.Current()
	if err != nil {
		t.Fatal(err)
	}
	if cur.Profile != "production" {
		t.Fatalf("profile = %q", cur.Profile)
	}
	if cur.ServerCount != 2 {
		t.Fatalf("servers = %d", cur.ServerCount)
	}
	target, err := os.Readlink(filepath.Join(st.BaseDir, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(target, "production.yaml") {
		t.Fatalf("symlink target = %q", target)
	}
}

func TestAddRejectsInlineToken(t *testing.T) {
	st := openTestStore(t)
	f, err := os.Open(filepath.Join("..", "..", "testdata", "inline-token.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := st.Add("bad", "", f, false); err == nil {
		t.Fatal("expected error")
	}
}

func TestRemoveActiveRequiresForce(t *testing.T) {
	st := openTestStore(t)
	if err := st.Add("production", "", readFixture(t), true); err != nil {
		t.Fatal(err)
	}
	if err := st.Remove("production", false); err == nil {
		t.Fatal("expected error")
	}
	if err := st.Remove("production", true); err != nil {
		t.Fatal(err)
	}
}

func TestListProfiles(t *testing.T) {
	st := openTestStore(t)
	if err := st.Add("production", "prod", readFixture(t), true); err != nil {
		t.Fatal(err)
	}
	profiles, current, err := st.ListProfiles()
	if err != nil {
		t.Fatal(err)
	}
	if current != "production" {
		t.Fatalf("current = %q", current)
	}
	if len(profiles) != 1 {
		t.Fatalf("profiles = %d", len(profiles))
	}
}

func TestValidateAll(t *testing.T) {
	st := openTestStore(t)
	if err := st.Add("production", "", readFixture(t), false); err != nil {
		t.Fatal(err)
	}
	warnings, err := st.ValidateProfile("all")
	if err != nil {
		t.Fatal(err)
	}
	_ = warnings
}

func TestTokenEnvsForProfile(t *testing.T) {
	st := openTestStore(t)
	if err := st.Add("production", "", readFixture(t), true); err != nil {
		t.Fatal(err)
	}
	envs, err := st.TokenEnvsForProfile("")
	if err != nil {
		t.Fatal(err)
	}
	if len(envs) != 2 {
		t.Fatalf("envs = %v", envs)
	}
}
