package paths_test

import (
	"testing"

	"github.com/nniksa91/cpanel-mcp-ctx/internal/paths"
)

func TestValidateProfileName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		valid   bool
	}{
		{"production", true},
		{"majofi-uapi", true},
		{"a", true},
		{"", false},
		{"../evil", false},
		{"state", false},
		{"UPPER", false},
	}
	for _, tc := range cases {
		err := paths.ValidateProfileName(tc.name)
		if tc.valid && err != nil {
			t.Fatalf("%q should be valid: %v", tc.name, err)
		}
		if !tc.valid && err == nil {
			t.Fatalf("%q should be invalid", tc.name)
		}
	}
}

func TestBaseDirOverride(t *testing.T) {
	t.Setenv("HOME", "/home/test")
	dir, err := paths.BaseDir("/tmp/custom")
	if err != nil {
		t.Fatal(err)
	}
	if dir != "/tmp/custom" {
		t.Fatalf("got %q", dir)
	}
}
