package command

import (
	"fmt"
	"io"
	"strings"

	"github.com/kincoy/cc9s/internal/cli/action"
	"github.com/kincoy/cc9s/internal/cli/contract"
	"github.com/kincoy/cc9s/internal/cli/render"
	"github.com/spf13/cobra"
)

var (
	runStatus = func() (contract.StatusResult, error) {
		return action.Status()
	}
	runProjectsList = func(opts contract.ProjectListOptions) (contract.ProjectListResult, error) {
		return action.ProjectsList(opts)
	}
	runProjectInspect = func(opts contract.ProjectInspectOptions) (contract.ProjectDetailResult, error) {
		return action.ProjectInspect(opts)
	}
	runSessionsList = func(opts contract.SessionListOptions) (contract.SessionListResult, error) {
		return action.SessionsList(opts)
	}
	runSessionInspect = func(opts contract.SessionInspectOptions) (contract.SessionDetailResult, error) {
		return action.SessionInspect(opts)
	}
	runSessionsCleanup = func(opts contract.CleanupOptions) (contract.CleanupResult, error) {
		return action.SessionsCleanup(opts)
	}
	runSkillsList = func(opts contract.SkillListOptions) (contract.SkillListResult, error) {
		return action.SkillsList(opts)
	}
	runAgentsList = func(opts contract.AgentListOptions) (contract.AgentListResult, error) {
		return action.AgentsList(opts)
	}
	runAgentInspect = func(opts contract.AgentInspectOptions) (contract.AgentDetailResult, error) {
		return action.AgentInspect(opts)
	}
	runThemes = func() contract.ThemesResult {
		return action.Themes()
	}
	runVersion = func() contract.VersionResult {
		return action.Version()
	}
)

type state struct {
	stdout  io.Writer
	stderr  io.Writer
	json    *bool
	version *bool
}

// New builds the Cobra command tree for the CLI.
func New(stdout, stderr io.Writer) *cobra.Command {
	jsonOutput := false
	versionOutput := false
	st := &state{
		stdout:  stdout,
		stderr:  stderr,
		json:    &jsonOutput,
		version: &versionOutput,
	}

	cmd := &cobra.Command{
		Use:           "cc9s",
		Short:         "Claude Code session manager",
		Long:          rootHelpLong(),
		Example:       rootHelpExamples(),
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if *st.version {
				return render.WriteResult(st.stdout, st.stderr, outputMode(*st.json), runVersion())
			}
			return fmt.Errorf("expected a command")
		},
	}

	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.PersistentFlags().BoolVar(st.json, "json", false, "Output JSON")
	cmd.PersistentFlags().BoolVarP(st.version, "version", "v", false, "Print version")

	cmd.AddCommand(
		newStatusCmd(st),
		newProjectsCmd(st),
		newSessionsCmd(st),
		newSkillsCmd(st),
		newAgentsCmd(st),
		newThemesCmd(st),
		newVersionCmd(st),
	)

	cmd.InitDefaultHelpCmd()

	return cmd
}

func newStatusCmd(st *state) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Environment health overview",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAndRender(st, func() (contract.StatusResult, error) {
				return runStatus()
			})
		},
	}
}

func newProjectsCmd(st *state) *cobra.Command {
	var opts contract.ProjectListOptions

	cmd := &cobra.Command{
		Use:     "projects",
		Aliases: []string{"project", "proj"},
		Short:   "List and inspect projects",
		Long:    projectsHelpLong(),
		Example: projectsHelpExamples(),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateProjectListOptions(opts); err != nil {
				return err
			}
			return runAndRender(st, func() (contract.ProjectListResult, error) {
				return runProjectsList(opts)
			})
		},
	}

	bindProjectListFlags(cmd, &opts)
	cmd.AddCommand(newProjectsListCmd(st), newProjectsInspectCmd(st))

	return cmd
}

func newProjectsListCmd(st *state) *cobra.Command {
	var opts contract.ProjectListOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateProjectListOptions(opts); err != nil {
				return err
			}
			return runAndRender(st, func() (contract.ProjectListResult, error) {
				return runProjectsList(opts)
			})
		},
	}

	bindProjectListFlags(cmd, &opts)
	return cmd
}

func newProjectsInspectCmd(st *state) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect <name>",
		Short: "Show project details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAndRender(st, func() (contract.ProjectDetailResult, error) {
				return runProjectInspect(contract.ProjectInspectOptions{Target: args[0]})
			})
		},
	}
	return cmd
}

func newSessionsCmd(st *state) *cobra.Command {
	var listOpts contract.SessionListOptions

	cmd := &cobra.Command{
		Use:     "sessions [id]",
		Aliases: []string{"session", "ss"},
		Short:   "List, inspect, and clean up sessions",
		Long:    sessionsHelpLong(),
		Example: sessionsHelpExamples(),
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				if err := validateImplicitSessionInspectFlags(cmd); err != nil {
					return err
				}
				return runAndRender(st, func() (contract.SessionDetailResult, error) {
					return runSessionInspect(contract.SessionInspectOptions{ID: args[0]})
				})
			}
			if err := validateSessionListOptions(listOpts); err != nil {
				return err
			}
			return runAndRender(st, func() (contract.SessionListResult, error) {
				return runSessionsList(listOpts)
			})
		},
	}

	bindSessionListFlags(cmd, &listOpts)
	cmd.AddCommand(newSessionsListCmd(st), newSessionsInspectCmd(st), newSessionsCleanupCmd(st))

	return cmd
}

func newSessionsListCmd(st *state) *cobra.Command {
	var opts contract.SessionListOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sessions across all projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateSessionListOptions(opts); err != nil {
				return err
			}
			return runAndRender(st, func() (contract.SessionListResult, error) {
				return runSessionsList(opts)
			})
		},
	}

	bindSessionListFlags(cmd, &opts)
	return cmd
}

func newSessionsInspectCmd(st *state) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect <id>",
		Short: "Show session details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAndRender(st, func() (contract.SessionDetailResult, error) {
				return runSessionInspect(contract.SessionInspectOptions{ID: args[0]})
			})
		},
	}
	return cmd
}

func newSessionsCleanupCmd(st *state) *cobra.Command {
	opts := contract.CleanupOptions{DryRun: false}

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Preview session cleanup candidates",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !opts.DryRun {
				return fmt.Errorf("--dry-run is required for cleanup (preview-only in v1)")
			}
			return runAndRender(st, func() (contract.CleanupResult, error) {
				return runSessionsCleanup(opts)
			})
		},
	}

	bindCleanupFlags(cmd, &opts)
	return cmd
}

func newSkillsCmd(st *state) *cobra.Command {
	var opts contract.SkillListOptions

	cmd := &cobra.Command{
		Use:     "skills",
		Aliases: []string{"skill", "sk"},
		Short:   "List skills and commands",
		Long:    skillsHelpLong(),
		Example: skillsHelpExamples(),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateSkillListOptions(opts); err != nil {
				return err
			}
			return runAndRender(st, func() (contract.SkillListResult, error) {
				return runSkillsList(opts)
			})
		},
	}

	bindSkillListFlags(cmd, &opts)
	cmd.AddCommand(newSkillsListCmd(st))

	return cmd
}

func newSkillsListCmd(st *state) *cobra.Command {
	var opts contract.SkillListOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List skills and commands",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateSkillListOptions(opts); err != nil {
				return err
			}
			return runAndRender(st, func() (contract.SkillListResult, error) {
				return runSkillsList(opts)
			})
		},
	}

	bindSkillListFlags(cmd, &opts)
	return cmd
}

func newAgentsCmd(st *state) *cobra.Command {
	var opts contract.AgentListOptions

	cmd := &cobra.Command{
		Use:     "agents",
		Aliases: []string{"agent", "ag"},
		Short:   "List and inspect agents",
		Long:    agentsHelpLong(),
		Example: agentsHelpExamples(),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateAgentListOptions(opts); err != nil {
				return err
			}
			return runAndRender(st, func() (contract.AgentListResult, error) {
				return runAgentsList(opts)
			})
		},
	}

	bindAgentListFlags(cmd, &opts)
	cmd.AddCommand(newAgentsListCmd(st), newAgentsInspectCmd(st))

	return cmd
}

func newAgentsListCmd(st *state) *cobra.Command {
	var opts contract.AgentListOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List agents",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateAgentListOptions(opts); err != nil {
				return err
			}
			return runAndRender(st, func() (contract.AgentListResult, error) {
				return runAgentsList(opts)
			})
		},
	}

	bindAgentListFlags(cmd, &opts)
	return cmd
}

func newAgentsInspectCmd(st *state) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect <name>",
		Short: "Show agent details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAndRender(st, func() (contract.AgentDetailResult, error) {
				return runAgentInspect(contract.AgentInspectOptions{Target: args[0]})
			})
		},
	}
	return cmd
}

func newThemesCmd(st *state) *cobra.Command {
	return &cobra.Command{
		Use:   "themes",
		Short: "List available themes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return render.WriteResult(st.stdout, st.stderr, outputMode(*st.json), runThemes())
		},
	}
}

func newVersionCmd(st *state) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return render.WriteResult(st.stdout, st.stderr, outputMode(*st.json), runVersion())
		},
	}
}

func outputMode(jsonOutput bool) contract.OutputMode {
	if jsonOutput {
		return contract.OutputJSON
	}
	return contract.OutputText
}

func runAndRender[T contract.Result](st *state, run func() (T, error)) error {
	result, err := run()
	if err != nil {
		return err
	}
	return render.WriteResult(st.stdout, st.stderr, outputMode(*st.json), result)
}

func bindProjectListFlags(cmd *cobra.Command, opts *contract.ProjectListOptions) {
	cmd.Flags().IntVar(&opts.Limit, "limit", 0, "Limit the number of results")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", "Sort field (name, sessions)")
}

func bindSessionListFlags(cmd *cobra.Command, opts *contract.SessionListOptions) {
	cmd.Flags().StringVar(&opts.Project, "project", "", "Filter by project name")
	cmd.Flags().StringVar(&opts.State, "state", "", "Filter by lifecycle state")
	cmd.Flags().IntVar(&opts.Limit, "limit", 0, "Limit the number of results")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", "Sort field (updated, state, project)")
}

func bindCleanupFlags(cmd *cobra.Command, opts *contract.CleanupOptions) {
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview-only cleanup mode")
	cmd.Flags().StringVar(&opts.Project, "project", "", "Filter by project name")
	cmd.Flags().StringVar(&opts.State, "state", "", "Filter by lifecycle state")
	cmd.Flags().StringVar(&opts.OlderThan, "older-than", "", "Only include sessions older than this duration")
}

func bindSkillListFlags(cmd *cobra.Command, opts *contract.SkillListOptions) {
	cmd.Flags().StringVar(&opts.Project, "project", "", "Filter by project name")
	cmd.Flags().StringVar(&opts.Scope, "scope", "", "Filter by scope")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by resource type")
}

func bindAgentListFlags(cmd *cobra.Command, opts *contract.AgentListOptions) {
	cmd.Flags().StringVar(&opts.Project, "project", "", "Filter by project name")
	cmd.Flags().StringVar(&opts.Scope, "scope", "", "Filter by scope")
}

func validateImplicitSessionInspectFlags(cmd *cobra.Command) error {
	for _, flag := range []string{"project", "state", "limit", "sort"} {
		if cmd.Flags().Changed(flag) {
			return fmt.Errorf("--%s is not supported for sessions inspect", flag)
		}
	}
	return nil
}

func validateProjectListOptions(opts contract.ProjectListOptions) error {
	switch opts.Sort {
	case "", "name", "sessions":
		return nil
	default:
		return fmt.Errorf("invalid --sort field %q (valid: name, sessions)", opts.Sort)
	}
}

func validateSessionListOptions(opts contract.SessionListOptions) error {
	switch opts.Sort {
	case "", "updated", "state", "project":
		return nil
	default:
		return fmt.Errorf("invalid --sort field %q (valid: updated, state, project)", opts.Sort)
	}
}

func validateSkillListOptions(opts contract.SkillListOptions) error {
	if opts.Scope == "" {
		return validateSkillType(opts.Type)
	}
	if !scopeMatches(opts.Scope) {
		return fmt.Errorf("invalid --scope value %q", opts.Scope)
	}
	return validateSkillType(opts.Type)
}

func validateAgentListOptions(opts contract.AgentListOptions) error {
	if opts.Scope == "" || scopeMatches(opts.Scope) {
		return nil
	}
	return fmt.Errorf("invalid --scope value %q", opts.Scope)
}

func validateSkillType(value string) error {
	if value == "" || strings.Contains("skill", strings.ToLower(strings.TrimSpace(value))) || strings.Contains(strings.ToLower(strings.TrimSpace(value)), "skill") ||
		strings.Contains("command", strings.ToLower(strings.TrimSpace(value))) || strings.Contains(strings.ToLower(strings.TrimSpace(value)), "command") {
		return nil
	}
	return fmt.Errorf("invalid --type value %q", value)
}

func scopeMatches(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return false
	}
	for _, candidate := range []string{"user", "project", "plugin"} {
		if strings.Contains(candidate, normalized) || strings.Contains(normalized, candidate) {
			return true
		}
	}
	return false
}

func rootHelpLong() string {
	return `Launch the interactive TUI when no command is provided.

Add --json for machine-readable output.
Use ` + "`cc9s <resource> --help`" + ` for resource-specific flags and enums.`
}

func rootHelpExamples() string {
	return `# Quick environment overview
cc9s status --json

# Find active sessions, then inspect one
cc9s sessions list --state active --json
cc9s sessions inspect <id> --json

# Preview cleanup candidates
cc9s sessions cleanup --dry-run --older-than 7d

# Inspect one project or list scoped resources
cc9s projects inspect cc9s
cc9s skills list --project cc9s --scope project --json`
}

func projectsHelpLong() string {
	return "Use `cc9s projects` or `cc9s projects list` to browse projects.\n" +
		"Use `cc9s projects inspect <name>` to inspect one project by display name or matching path.\n\n" +
		"Output contract:\n" +
		"  list -> JSON array\n" +
		"  inspect -> JSON object\n\n" +
		"Sort values:\n" +
		"  --sort: name | sessions"
}

func projectsHelpExamples() string {
	return `# List projects sorted by session count
cc9s projects list --sort sessions --json

# Inspect one project by name
cc9s projects inspect cc9s

# Inspect one project by matching path
cc9s projects inspect /Users/kinco/go/src/github.com/kincoy/cc9s`
}

func sessionsHelpLong() string {
	return "Use `cc9s sessions` or `cc9s sessions list` to browse sessions across projects.\n" +
		"`cc9s sessions <id>` is an inspect shortcut.\n" +
		"`cleanup` is preview-only and requires `--dry-run`.\n\n" +
		"Enums:\n" +
		"  --state: active | idle | completed | stale\n" +
		"  --sort: updated | state | project\n\n" +
		"Output contract:\n" +
		"  list -> JSON array\n" +
		"  inspect / cleanup -> JSON object"
}

func sessionsHelpExamples() string {
	return `# Find active sessions and capture IDs
cc9s sessions list --state active --json

# Inspect one session in detail
cc9s sessions inspect <id> --json
cc9s sessions <id> --json

# Preview cleanup candidates
cc9s sessions cleanup --dry-run --older-than 7d

# Filter sessions for one project
cc9s sessions list --project cc9s --sort updated`
}

func skillsHelpLong() string {
	return "List discovered skills and commands from user, project, and plugin scopes.\n\n" +
		"Enums:\n" +
		"  --scope: user | project | plugin\n" +
		"  --type: skill | command\n\n" +
		"Output contract:\n" +
		"  list -> JSON array"
}

func skillsHelpExamples() string {
	return `# List all available skills and commands
cc9s skills list

# Narrow to one project and one scope
cc9s skills list --project cc9s --scope project --json

# Show only skill entries
cc9s skills list --type skill`
}

func agentsHelpLong() string {
	return "List file-backed agents across user, project, and plugin scopes, or inspect one agent by name/path.\n\n" +
		"Enums:\n" +
		"  --scope: user | project | plugin\n\n" +
		"Output contract:\n" +
		"  list -> JSON array\n" +
		"  inspect -> JSON object"
}

func agentsHelpExamples() string {
	return `# List all discovered agents
cc9s agents list --json

# Narrow to one scope
cc9s agents list --scope plugin

# Inspect one agent
cc9s agents inspect reviewer`
}
