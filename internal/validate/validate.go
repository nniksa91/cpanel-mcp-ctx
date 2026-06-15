package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	dirPerm  = 0o700
	filePerm = 0o600
)

// ConfigSummary is a redacted view of a parsed MCP config file.
type ConfigSummary struct {
	DefaultServer string
	Servers       []ServerSummary
}

type ServerSummary struct {
	Name     string
	Type     string
	Host     string
	Port     int
	Username string
	TokenEnv string
}

// ValidateFile parses and validates a cPanel MCP YAML config at path.
func ValidateFile(path string) (*ConfigSummary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return ValidateBytes(data, path)
}

// ValidateBytes validates raw YAML content.
func ValidateBytes(data []byte, source string) (*ConfigSummary, error) {
	if source == "" {
		source = "config"
	}
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("YAML error in %s: %w", source, err)
	}
	if raw == nil {
		return nil, fmt.Errorf("%s: top-level config must be a mapping", source)
	}
	return parseConfig(raw, source)
}

func parseConfig(raw map[string]any, source string) (*ConfigSummary, error) {
	serversRaw, ok := raw["servers"]
	if !ok {
		return nil, fmt.Errorf("%s: `servers` must be a non-empty mapping", source)
	}
	serversMap, ok := serversRaw.(map[string]any)
	if !ok || len(serversMap) == 0 {
		return nil, fmt.Errorf("%s: `servers` must be a non-empty mapping", source)
	}

	summary := &ConfigSummary{Servers: make([]ServerSummary, 0, len(serversMap))}
	names := make(map[string]struct{}, len(serversMap))

	for name, body := range serversMap {
		if err := validateProfileSegment(name); err != nil {
			return nil, fmt.Errorf("%s: server name %q: %w", source, name, err)
		}
		serverBody, ok := body.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s: server %q must be a mapping", source, name)
		}
		s, err := parseServer(name, serverBody, source)
		if err != nil {
			return nil, err
		}
		summary.Servers = append(summary.Servers, s)
		names[name] = struct{}{}
	}

	defaultServer, _ := raw["default_server"].(string)
	if defaultServer != "" {
		if _, ok := names[defaultServer]; !ok {
			return nil, fmt.Errorf(
				"%s: default_server %q is not in `servers`",
				source, defaultServer,
			)
		}
		summary.DefaultServer = defaultServer
	} else if len(summary.Servers) == 1 {
		summary.DefaultServer = summary.Servers[0].Name
	}

	return summary, nil
}

func parseServer(name string, body map[string]any, source string) (ServerSummary, error) {
	apiType, _ := body["type"].(string)
	if apiType != "whm" && apiType != "uapi" {
		return ServerSummary{}, fmt.Errorf(
			"%s: server %q `type` must be 'whm' or 'uapi', got %q",
			source, name, apiType,
		)
	}

	host, _ := body["host"].(string)
	host = strings.TrimSpace(host)
	if host == "" {
		return ServerSummary{}, fmt.Errorf("%s: server %q `host` is required", source, name)
	}
	if strings.Contains(host, "://") {
		return ServerSummary{}, fmt.Errorf(
			"%s: server %q `host` must be a hostname, not a URL",
			source, name,
		)
	}

	port := defaultPort(apiType)
	if rawPort, ok := body["port"]; ok {
		switch v := rawPort.(type) {
		case int:
			port = v
		case int64:
			port = int(v)
		default:
			return ServerSummary{}, fmt.Errorf(
				"%s: server %q `port` is invalid: %v",
				source, name, rawPort,
			)
		}
	}
	if port <= 0 || port > 65535 {
		return ServerSummary{}, fmt.Errorf("%s: server %q `port` is invalid: %d", source, name, port)
	}

	authRaw, ok := body["auth"].(map[string]any)
	if !ok {
		return ServerSummary{}, fmt.Errorf("%s: server %q `auth` is required", source, name)
	}
	username, tokenEnv, err := parseAuth(name, authRaw, source)
	if err != nil {
		return ServerSummary{}, err
	}

	if err := checkOperationLists(name, body, source); err != nil {
		return ServerSummary{}, err
	}

	return ServerSummary{
		Name:     name,
		Type:     apiType,
		Host:     host,
		Port:     port,
		Username: username,
		TokenEnv: tokenEnv,
	}, nil
}

func parseAuth(serverName string, body map[string]any, source string) (username, tokenEnv string, err error) {
	method, _ := body["method"].(string)
	if method == "" {
		method = "token"
	}
	if method != "token" {
		return "", "", fmt.Errorf(
			"%s: server %q auth.method must be 'token' (got %q)",
			source, serverName, method,
		)
	}
	user, _ := body["username"].(string)
	user = strings.TrimSpace(user)
	if user == "" {
		return "", "", fmt.Errorf("%s: server %q auth.username is required", source, serverName)
	}

	_, hasToken := body["token"]
	te, hasTokenEnv := body["token_env"].(string)
	if hasToken && hasTokenEnv {
		return "", "", fmt.Errorf(
			"%s: server %q auth: set either token_env or token, not both",
			source, serverName,
		)
	}
	if hasToken {
		return "", "", fmt.Errorf(
			"%s: server %q auth: inline `token:` is not allowed; use `token_env:` only",
			source, serverName,
		)
	}
	if !hasTokenEnv || strings.TrimSpace(te) == "" {
		return "", "", fmt.Errorf(
			"%s: server %q auth: token_env is required",
			source, serverName,
		)
	}
	return user, strings.TrimSpace(te), nil
}

func checkOperationLists(serverName string, body map[string]any, source string) error {
	allowed := opSet(body["allowed_operations"])
	denied := opSet(body["denied_operations"])
	for k := range allowed {
		if _, ok := denied[k]; ok {
			return fmt.Errorf(
				"%s: server %q lists the same operation(s) in both allowed_operations and denied_operations: %s",
				source, serverName, k,
			)
		}
	}
	return nil
}

func opSet(raw any) map[string]struct{} {
	out := map[string]struct{}{}
	list, ok := raw.([]any)
	if !ok {
		return out
	}
	for _, item := range list {
		s, ok := item.(string)
		if !ok {
			continue
		}
		out[strings.ToLower(strings.TrimSpace(s))] = struct{}{}
	}
	return out
}

func defaultPort(apiType string) int {
	if apiType == "uapi" {
		return 2083
	}
	return 2087
}

func validateProfileSegment(name string) error {
	if strings.Contains(name, "/") || strings.Contains(name, `\`) {
		return fmt.Errorf("must not contain path separators")
	}
	return nil
}

// CheckPermissions validates directory and file modes under baseDir.
func CheckPermissions(baseDir string) []string {
	var warnings []string
	check := func(path string, want os.FileMode, isDir bool) {
		info, err := os.Lstat(path)
		if err != nil {
			return
		}
		mode := info.Mode().Perm()
		if isDir && mode&0o077 != 0 {
			warnings = append(warnings, fmt.Sprintf("%s: directory mode %04o should be %04o", path, mode, want))
		}
		if !isDir && mode&0o077 != 0 {
			warnings = append(warnings, fmt.Sprintf("%s: file mode %04o should be %04o", path, mode, want))
		}
	}
	check(baseDir, dirPerm, true)
	check(filepath.Join(baseDir, "configs"), dirPerm, true)
	return warnings
}

// TokenEnvs returns unique token_env names from a config summary.
func TokenEnvs(summary *ConfigSummary) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range summary.Servers {
		if s.TokenEnv == "" {
			continue
		}
		if _, ok := seen[s.TokenEnv]; ok {
			continue
		}
		seen[s.TokenEnv] = struct{}{}
		out = append(out, s.TokenEnv)
	}
	return out
}
