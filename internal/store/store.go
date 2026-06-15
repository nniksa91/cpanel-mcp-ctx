package store

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nniksa91/cpanel-mcp-ctx/internal/paths"
	"github.com/nniksa91/cpanel-mcp-ctx/internal/validate"
	"gopkg.in/yaml.v3"
)

const (
	dirPerm  = 0o700
	filePerm = 0o600
)

// State persisted alongside profile configs.
type State struct {
	Version  int                       `yaml:"version"`
	Current  string                    `yaml:"current"`
	Profiles map[string]ProfileMeta    `yaml:"profiles"`
}

type ProfileMeta struct {
	Description string    `yaml:"description,omitempty"`
	CreatedAt   time.Time `yaml:"created_at"`
}

// Store manages ~/.cpanel-mcp profile files.
type Store struct {
	BaseDir string
}

func Open(baseDir string) (*Store, error) {
	dir, err := paths.BaseDir(baseDir)
	if err != nil {
		return nil, err
	}
	return &Store{BaseDir: dir}, nil
}

func (s *Store) Initialized() bool {
	info, err := os.Stat(s.BaseDir)
	return err == nil && info.IsDir()
}

// Init creates the directory tree with safe permissions.
func (s *Store) Init(fixPermissions bool) error {
	if err := os.MkdirAll(paths.ConfigsDir(s.BaseDir), dirPerm); err != nil {
		return fmt.Errorf("create configs dir: %w", err)
	}
	if err := os.Chmod(s.BaseDir, dirPerm); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("chmod base dir: %w", err)
	}
	if err := os.Chmod(paths.ConfigsDir(s.BaseDir), dirPerm); err != nil {
		return fmt.Errorf("chmod configs dir: %w", err)
	}
	state, err := s.loadState()
	if err != nil {
		return err
	}
	if state.Profiles == nil {
		state = State{Version: 1, Profiles: map[string]ProfileMeta{}}
		if err := s.saveState(state); err != nil {
			return err
		}
	}
	if fixPermissions {
		return s.fixPermissions()
	}
	return nil
}

func (s *Store) fixPermissions() error {
	return filepath.Walk(s.BaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return os.Chmod(path, dirPerm)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		return os.Chmod(path, filePerm)
	})
}

func (s *Store) loadState() (State, error) {
	path := paths.StateFile(s.BaseDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return State{Version: 1, Profiles: map[string]ProfileMeta{}}, nil
		}
		return State{}, fmt.Errorf("read state: %w", err)
	}
	var st State
	if err := yaml.Unmarshal(data, &st); err != nil {
		return State{}, fmt.Errorf("parse state.yaml: %w", err)
	}
	if st.Profiles == nil {
		st.Profiles = map[string]ProfileMeta{}
	}
	if st.Version == 0 {
		st.Version = 1
	}
	return st, nil
}

func (s *Store) saveState(st State) error {
	if st.Profiles == nil {
		st.Profiles = map[string]ProfileMeta{}
	}
	if st.Version == 0 {
		st.Version = 1
	}
	data, err := yaml.Marshal(st)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	return writeFileAtomic(paths.StateFile(s.BaseDir), data, filePerm)
}

// ProfileInfo describes one installed profile.
type ProfileInfo struct {
	Name          string
	Active        bool
	ConfigPath    string
	Description   string
	ServerCount   int
	DefaultServer string
}

func (s *Store) ListProfiles() ([]ProfileInfo, string, error) {
	st, err := s.loadState()
	if err != nil {
		return nil, "", err
	}
	current := s.resolveCurrent(st)

	entries, err := os.ReadDir(paths.ConfigsDir(s.BaseDir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, current, nil
		}
		return nil, "", fmt.Errorf("list configs: %w", err)
	}

	var profiles []ProfileInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".yaml")
		path := paths.ProfileFile(s.BaseDir, name)
		info := ProfileInfo{
			Name:       name,
			Active:     name == current,
			ConfigPath: path,
		}
		if meta, ok := st.Profiles[name]; ok {
			info.Description = meta.Description
		}
		if summary, err := validate.ValidateFile(path); err == nil {
			info.ServerCount = len(summary.Servers)
			info.DefaultServer = summary.DefaultServer
		}
		profiles = append(profiles, info)
	}
	return profiles, current, nil
}

type CurrentInfo struct {
	Profile      string
	ConfigPath   string
	ResolvedPath string
	DefaultServer string
	ServerCount  int
	SymlinkOK    bool
}

func (s *Store) Current() (CurrentInfo, error) {
	st, err := s.loadState()
	if err != nil {
		return CurrentInfo{}, err
	}
	current := s.resolveCurrent(st)
	info := CurrentInfo{
		Profile:    current,
		ConfigPath: paths.ActiveLink(s.BaseDir),
	}
	if current == "" {
		return info, nil
	}
	info.ResolvedPath = paths.ProfileFile(s.BaseDir, current)
	if target, err := os.Readlink(info.ConfigPath); err == nil {
		info.SymlinkOK = filepath.Clean(target) == filepath.Join("configs", current+".yaml") ||
			filepath.Clean(target) == info.ResolvedPath
	}
	if summary, err := validate.ValidateFile(info.ResolvedPath); err == nil {
		info.DefaultServer = summary.DefaultServer
		info.ServerCount = len(summary.Servers)
	}
	return info, nil
}

func (s *Store) resolveCurrent(st State) string {
	if st.Current != "" {
		if _, err := os.Stat(paths.ProfileFile(s.BaseDir, st.Current)); err == nil {
			return st.Current
		}
	}
	if target, err := os.Readlink(paths.ActiveLink(s.BaseDir)); err == nil {
		base := filepath.Base(target)
		if strings.HasSuffix(base, ".yaml") {
			return strings.TrimSuffix(base, ".yaml")
		}
	}
	return st.Current
}

// Use switches the active profile.
func (s *Store) Use(profile string) error {
	if err := paths.ValidateProfileName(profile); err != nil {
		return err
	}
	target := paths.ProfileFile(s.BaseDir, profile)
	if _, err := os.Stat(target); err != nil {
		return fmt.Errorf("profile %q not found; run `cpanel-mcp-ctx list`", profile)
	}
	if _, err := validate.ValidateFile(target); err != nil {
		return err
	}
	if err := s.setActiveSymlink(profile); err != nil {
		return err
	}
	st, err := s.loadState()
	if err != nil {
		return err
	}
	st.Current = profile
	if st.Profiles == nil {
		st.Profiles = map[string]ProfileMeta{}
	}
	if _, ok := st.Profiles[profile]; !ok {
		st.Profiles[profile] = ProfileMeta{CreatedAt: time.Now().UTC()}
	}
	return s.saveState(st)
}

func (s *Store) setActiveSymlink(profile string) error {
	link := paths.ActiveLink(s.BaseDir)
	target := filepath.Join("configs", profile+".yaml")
	temp := link + ".new"
	_ = os.Remove(temp)
	if err := os.Symlink(target, temp); err != nil {
		return fmt.Errorf("create symlink: %w", err)
	}
	if err := os.Rename(temp, link); err != nil {
		_ = os.Remove(temp)
		return fmt.Errorf("activate symlink: %w", err)
	}
	return nil
}

// Add installs a profile from reader.
func (s *Store) Add(profile, description string, src io.Reader, setCurrent bool) error {
	if err := paths.ValidateProfileName(profile); err != nil {
		return err
	}
	if !s.Initialized() {
		if err := s.Init(false); err != nil {
			return err
		}
	}
	data, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("read source config: %w", err)
	}
	dest := paths.ProfileFile(s.BaseDir, profile)
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("profile %q already exists", profile)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat profile: %w", err)
	}
	if _, err := validate.ValidateBytes(data, dest); err != nil {
		return err
	}
	if err := writeFileAtomic(dest, data, filePerm); err != nil {
		return err
	}
	st, err := s.loadState()
	if err != nil {
		return err
	}
	if st.Profiles == nil {
		st.Profiles = map[string]ProfileMeta{}
	}
	st.Profiles[profile] = ProfileMeta{
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.saveState(st); err != nil {
		return err
	}
	if setCurrent {
		return s.Use(profile)
	}
	return nil
}

// Remove deletes a profile.
func (s *Store) Remove(profile string, force bool) error {
	if err := paths.ValidateProfileName(profile); err != nil {
		return err
	}
	st, err := s.loadState()
	if err != nil {
		return err
	}
	current := s.resolveCurrent(st)
	if current == profile && !force {
		return fmt.Errorf(
			"profile %q is active; switch away first or use --force",
			profile,
		)
	}
	path := paths.ProfileFile(s.BaseDir, profile)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove profile: %w", err)
	}
	delete(st.Profiles, profile)
	if st.Current == profile {
		st.Current = ""
		_ = os.Remove(paths.ActiveLink(s.BaseDir))
	}
	return s.saveState(st)
}

// ValidateProfile validates one profile or all profiles.
func (s *Store) ValidateProfile(profile string) ([]string, error) {
	if profile == "all" {
		entries, err := os.ReadDir(paths.ConfigsDir(s.BaseDir))
		if err != nil {
			return nil, err
		}
		var issues []string
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			name := strings.TrimSuffix(e.Name(), ".yaml")
			if msgs, err := s.validateOne(name); err != nil {
				return nil, err
			} else {
				issues = append(issues, msgs...)
			}
		}
		return issues, nil
	}
	if profile == "" {
		cur, err := s.Current()
		if err != nil {
			return nil, err
		}
		if cur.Profile == "" {
			return []string{"no active profile"}, nil
		}
		profile = cur.Profile
	}
	return s.validateOne(profile)
}

func (s *Store) validateOne(profile string) ([]string, error) {
	path := paths.ProfileFile(s.BaseDir, profile)
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("profile %q not found", profile)
	}
	if _, err := validate.ValidateFile(path); err != nil {
		return nil, err
	}
	var msgs []string
	msgs = append(msgs, validate.CheckPermissions(s.BaseDir)...)
	return msgs, nil
}

// ShowSummary returns a validated summary for a profile.
func (s *Store) ShowSummary(profile string) (*validate.ConfigSummary, string, error) {
	if profile == "" {
		cur, err := s.Current()
		if err != nil {
			return nil, "", err
		}
		if cur.Profile == "" {
			return nil, "", fmt.Errorf("no active profile")
		}
		profile = cur.Profile
	}
	path := paths.ProfileFile(s.BaseDir, profile)
	summary, err := validate.ValidateFile(path)
	if err != nil {
		return nil, "", err
	}
	return summary, path, nil
}

// TokenEnvsForProfile lists token_env names for a profile.
func (s *Store) TokenEnvsForProfile(profile string) ([]string, error) {
	summary, _, err := s.ShowSummary(profile)
	if err != nil {
		return nil, err
	}
	return validate.TokenEnvs(summary), nil
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Chmod(perm); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempName, path)
}
