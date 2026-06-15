package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/nniksa91/cpanel-mcp-ctx/internal/store"
	"github.com/nniksa91/cpanel-mcp-ctx/internal/validate"
)

type Options struct {
	JSON  bool
	Quiet bool
	Out   io.Writer
	Err   io.Writer
}

func (o Options) stdout() io.Writer {
	if o.Out != nil {
		return o.Out
	}
	return os.Stdout
}

func (o Options) stderr() io.Writer {
	if o.Err != nil {
		return o.Err
	}
	return os.Stderr
}

func (o Options) Printf(format string, args ...any) {
	if o.Quiet {
		return
	}
	fmt.Fprintf(o.stdout(), format, args...)
}

func (o Options) Println(args ...any) {
	if o.Quiet {
		return
	}
	fmt.Fprintln(o.stdout(), args...)
}

func (o Options) Note(msg string) {
	if o.Quiet {
		return
	}
	fmt.Fprintln(o.stderr(), msg)
}

func (o Options) WriteJSON(v any) error {
	enc := json.NewEncoder(o.stdout())
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func (o Options) PrintList(profiles []store.ProfileInfo, current string) error {
	if o.JSON {
		type row struct {
			Profile       string `json:"profile"`
			Active        bool   `json:"active"`
			ServerCount   int    `json:"server_count"`
			DefaultServer string `json:"default_server,omitempty"`
			Description   string `json:"description,omitempty"`
		}
		rows := make([]row, 0, len(profiles))
		for _, p := range profiles {
			rows = append(rows, row{
				Profile:       p.Name,
				Active:        p.Active,
				ServerCount:   p.ServerCount,
				DefaultServer: p.DefaultServer,
				Description:   p.Description,
			})
		}
		return o.WriteJSON(map[string]any{
			"current":  current,
			"profiles": rows,
		})
	}
	w := tabwriter.NewWriter(o.stdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PROFILE\tACTIVE\tSERVERS\tDEFAULT_SERVER\tDESCRIPTION")
	for _, p := range profiles {
		active := ""
		if p.Active {
			active = "*"
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
			p.Name, active, p.ServerCount, p.DefaultServer, p.Description)
	}
	return w.Flush()
}

func (o Options) PrintCurrent(info store.CurrentInfo) error {
	if o.JSON {
		return o.WriteJSON(map[string]any{
			"profile":        info.Profile,
			"config_path":    info.ConfigPath,
			"resolved_path":  info.ResolvedPath,
			"default_server": info.DefaultServer,
			"server_count":   info.ServerCount,
			"symlink_ok":     info.SymlinkOK,
		})
	}
	if info.Profile == "" {
		o.Println("no active profile")
		return nil
	}
	o.Printf("profile: %s\n", info.Profile)
	o.Printf("config: %s\n", info.ConfigPath)
	o.Printf("resolved: %s\n", info.ResolvedPath)
	o.Printf("default_server: %s\n", info.DefaultServer)
	o.Printf("servers: %d\n", info.ServerCount)
	if !info.SymlinkOK {
		o.Note("warning: active symlink does not match state.yaml current profile")
	}
	return nil
}

func (o Options) PrintShow(profile string, summary *validate.ConfigSummary, path string) error {
	if o.JSON {
		servers := make([]map[string]any, 0, len(summary.Servers))
		for _, s := range summary.Servers {
			servers = append(servers, map[string]any{
				"name":      s.Name,
				"type":      s.Type,
				"host":      s.Host,
				"port":      s.Port,
				"username":  s.Username,
				"token_env": s.TokenEnv,
			})
		}
		return o.WriteJSON(map[string]any{
			"profile":        profile,
			"path":           path,
			"default_server": summary.DefaultServer,
			"servers":        servers,
		})
	}
	o.Printf("profile: %s\n", profile)
	o.Printf("path: %s\n", path)
	o.Printf("default_server: %s\n", summary.DefaultServer)
	o.Println("servers:")
	for _, s := range summary.Servers {
		o.Printf("  - %s (%s) %s:%d user=%s token_env=%s\n",
			s.Name, s.Type, s.Host, s.Port, s.Username, s.TokenEnv)
	}
	return nil
}

func (o Options) PrintEnvExports(names []string, check bool) error {
	if o.JSON {
		status := map[string]string{}
		for _, n := range names {
			if v, ok := os.LookupEnv(n); ok && v != "" {
				status[n] = "set"
			} else {
				status[n] = "unset"
			}
		}
		return o.WriteJSON(map[string]any{"token_env": status})
	}
	missing := []string{}
	for _, n := range names {
		v, ok := os.LookupEnv(n)
		if !ok || v == "" {
			missing = append(missing, n)
			continue
		}
		if !check {
			fmt.Fprintf(o.stdout(), "export %s=%q\n", n, v)
		}
	}
	if check && len(missing) > 0 {
		return fmt.Errorf("unset token_env variable(s): %s", strings.Join(missing, ", "))
	}
	return nil
}

func (o Options) PrintValidateOK(profile string, warnings []string) {
	if o.JSON {
		_ = o.WriteJSON(map[string]any{"profile": profile, "ok": true, "warnings": warnings})
		return
	}
	o.Printf("profile %q is valid\n", profile)
	for _, w := range warnings {
		o.Note("warning: " + w)
	}
}
