package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	EnvDirOverride = "CPANEL_MCP_CTX_DIR"
	DefaultBase    = ".cpanel-mcp"
)

var profileNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

var reservedProfiles = map[string]struct{}{
	"state": {}, "config": {}, "configs": {}, ".": {}, "..": {},
}

// BaseDir resolves the cpanel-mcp state directory (~/.cpanel-mcp by default).
func BaseDir(override string) (string, error) {
	if override != "" {
		return filepath.Abs(expandHome(override))
	}
	if v := strings.TrimSpace(os.Getenv(EnvDirOverride)); v != "" {
		return filepath.Abs(expandHome(v))
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home directory: %w", err)
	}
	return filepath.Join(home, DefaultBase), nil
}

func expandHome(p string) string {
	if p == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return p
	}
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

// ValidateProfileName checks profile identifiers for safe filesystem use.
func ValidateProfileName(name string) error {
	if name == "" {
		return fmt.Errorf("profile name is required")
	}
	if _, reserved := reservedProfiles[name]; reserved {
		return fmt.Errorf("profile name %q is reserved", name)
	}
	if !profileNameRe.MatchString(name) {
		return fmt.Errorf(
			"profile name %q is invalid; use [a-z0-9][a-z0-9._-]{0,63}",
			name,
		)
	}
	return nil
}

// Layout paths under baseDir.
func ConfigsDir(baseDir string) string  { return filepath.Join(baseDir, "configs") }
func StateFile(baseDir string) string    { return filepath.Join(baseDir, "state.yaml") }
func ActiveLink(baseDir string) string     { return filepath.Join(baseDir, "config.yaml") }
func ProfileFile(baseDir, profile string) string {
	return filepath.Join(ConfigsDir(baseDir), profile+".yaml")
}
