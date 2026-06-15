package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/nniksa91/cpanel-mcp-ctx/internal/output"
	"github.com/nniksa91/cpanel-mcp-ctx/internal/store"
	"github.com/spf13/cobra"
)

const restartHint = "cpanel-mcp reloads the active profile on the next tool call; restart MCP only if tools still show the old profile."

var (
	baseDir string
	jsonOut bool
	quiet   bool
)

// Execute runs the root command.
func Execute(version string) error {
	root := newRoot(version)
	return root.Execute()
}

func newRoot(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "cpanel-mcp-ctx",
		Short: "Switch cPanel MCP configuration profiles",
		Long: strings.TrimSpace(`
Manage multiple cPanel MCP YAML configs under ~/.cpanel-mcp/configs/ and
switch the active profile via a symlink read by cpanel-mcp.

Works with any MCP host (Cursor, Claude Desktop, VS Code, etc.) that launches
cpanel-mcp with --config or the default ~/.cpanel-mcp/config.yaml path.
`),
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:      version,
	}
	root.PersistentFlags().StringVar(&baseDir, "cpanel-mcp-dir", "", "override ~/.cpanel-mcp state directory")
	root.PersistentFlags().BoolVar(&jsonOut, "json", false, "emit JSON output")
	root.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")

	root.AddCommand(
		cmdInit(),
		cmdList(),
		cmdCurrent(),
		cmdUse(),
		cmdAdd(),
		cmdRemove(),
		cmdShow(),
		cmdValidate(),
		cmdEnv(),
	)
	return root
}

func openStore() (*store.Store, output.Options, error) {
	st, err := store.Open(baseDir)
	if err != nil {
		return nil, output.Options{}, err
	}
	return st, output.Options{JSON: jsonOut, Quiet: quiet}, nil
}

func cmdInit() *cobra.Command {
	fix := false
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create ~/.cpanel-mcp directory tree",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, out, err := openStore()
			if err != nil {
				return err
			}
			if err := st.Init(fix); err != nil {
				return err
			}
			if out.JSON {
				return out.WriteJSON(map[string]any{"initialized": true, "base_dir": st.BaseDir})
			}
			out.Printf("initialized %s\n", st.BaseDir)
			return nil
		},
	}
	cmd.Flags().BoolVar(&fix, "fix-permissions", false, "repair file modes to 0700/0600")
	return cmd
}

func cmdList() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, out, err := openStore()
			if err != nil {
				return err
			}
			profiles, current, err := st.ListProfiles()
			if err != nil {
				return err
			}
			return out.PrintList(profiles, current)
		},
	}
}

func cmdCurrent() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the active profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, out, err := openStore()
			if err != nil {
				return err
			}
			info, err := st.Current()
			if err != nil {
				return err
			}
			return out.PrintCurrent(info)
		},
	}
}

func cmdUse() *cobra.Command {
	return &cobra.Command{
		Use:   "use <profile>",
		Short: "Switch to a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, out, err := openStore()
			if err != nil {
				return err
			}
			if !st.Initialized() {
				return fmt.Errorf("not initialized; run `cpanel-mcp-ctx init` first")
			}
			if err := st.Use(args[0]); err != nil {
				return err
			}
			cur, err := st.Current()
			if err != nil {
				return err
			}
			if out.JSON {
				if err := out.WriteJSON(map[string]any{
					"profile":       cur.Profile,
					"resolved_path": cur.ResolvedPath,
					"server_count":  cur.ServerCount,
				}); err != nil {
					return err
				}
			} else {
				out.Printf("Switched to profile %q (%s, %d servers)\n",
					cur.Profile, cur.ResolvedPath, cur.ServerCount)
			}
			out.Note(restartHint)
			return nil
		},
	}
}

func cmdAdd() *cobra.Command {
	var from string
	var description string
	var setCurrent bool
	cmd := &cobra.Command{
		Use:   "add <profile>",
		Short: "Install a profile from a YAML file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, out, err := openStore()
			if err != nil {
				return err
			}
			if from == "" {
				return fmt.Errorf("--from is required")
			}
			var src *os.File
			if from == "-" {
				src = os.Stdin
			} else {
				src, err = os.Open(from)
				if err != nil {
					return fmt.Errorf("open source: %w", err)
				}
				defer src.Close()
			}
			if err := st.Add(args[0], description, src, setCurrent); err != nil {
				return err
			}
			path := fmt.Sprintf("%s/configs/%s.yaml", st.BaseDir, args[0])
			if out.JSON {
				return out.WriteJSON(map[string]any{
					"profile": args[0],
					"path":    path,
					"active":  setCurrent,
				})
			}
			out.Printf("Added profile %q at %s\n", args[0], path)
			if setCurrent {
				out.Note(restartHint)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "source YAML file (- for stdin)")
	cmd.Flags().StringVar(&description, "description", "", "profile description stored in state.yaml")
	cmd.Flags().BoolVar(&setCurrent, "set-current", false, "activate profile after add")
	return cmd
}

func cmdRemove() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "remove <profile>",
		Short: "Delete a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, out, err := openStore()
			if err != nil {
				return err
			}
			if err := st.Remove(args[0], force); err != nil {
				return err
			}
			if out.JSON {
				return out.WriteJSON(map[string]any{"removed": args[0]})
			}
			out.Printf("Removed profile %q\n", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "remove even if active")
	return cmd
}

func cmdShow() *cobra.Command {
	return &cobra.Command{
		Use:   "show [profile]",
		Short: "Show a redacted profile summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, out, err := openStore()
			if err != nil {
				return err
			}
			profile := ""
			if len(args) > 0 {
				profile = args[0]
			}
			summary, path, err := st.ShowSummary(profile)
			if err != nil {
				return err
			}
			if profile == "" {
				cur, err := st.Current()
				if err != nil {
					return err
				}
				profile = cur.Profile
			}
			return out.PrintShow(profile, summary, path)
		},
	}
}

func cmdValidate() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [profile|all]",
		Short: "Validate profile YAML and permissions",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, out, err := openStore()
			if err != nil {
				return err
			}
			target := ""
			if len(args) > 0 {
				target = args[0]
			}
			warnings, err := st.ValidateProfile(target)
			if err != nil {
				return err
			}
			name := target
			if name == "" {
				cur, _ := st.Current()
				name = cur.Profile
			}
			if target == "all" {
				name = "all"
			} else if name == "" {
				name = "active"
			}
			out.PrintValidateOK(name, warnings)
			return nil
		},
	}
}

func cmdEnv() *cobra.Command {
	var check bool
	cmd := &cobra.Command{
		Use:   "env [profile]",
		Short: "Print or check token_env variables for a profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, out, err := openStore()
			if err != nil {
				return err
			}
			profile := ""
			if len(args) > 0 {
				profile = args[0]
			}
			names, err := st.TokenEnvsForProfile(profile)
			if err != nil {
				return err
			}
			if len(names) == 0 {
				return fmt.Errorf("no token_env references found")
			}
			if err := out.PrintEnvExports(names, check); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "exit non-zero if any token_env is unset")
	return cmd
}
