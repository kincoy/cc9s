package render

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/kincoy/cc9s/internal/cli/contract"
)

func writeTextResult(stdout, stderr io.Writer, result contract.Result) error {
	switch r := result.(type) {
	case contract.StatusResult:
		return writeString(stdout, renderStatusText(r))
	case contract.ProjectListResult:
		return writeString(stdout, renderProjectListText(r))
	case contract.ProjectDetailResult:
		return writeString(stdout, renderProjectDetailText(r))
	case contract.SessionListResult:
		return writeString(stdout, renderSessionListText(r))
	case contract.SessionDetailResult:
		return writeString(stdout, renderSessionDetailText(r))
	case contract.SkillListResult:
		return writeString(stdout, renderSkillListText(r))
	case contract.AgentListResult:
		return writeString(stdout, renderAgentListText(r))
	case contract.AgentDetailResult:
		return writeString(stdout, renderAgentDetailText(r))
	case contract.ThemesResult:
		return writeString(stdout, renderThemesText(r))
	case contract.CleanupResult:
		return writeString(stdout, renderCleanupText(r))
	case contract.VersionResult:
		return writeString(stdout, fmt.Sprintf("cc9s v%s\n", r.Version))
	default:
		return writeString(stdout, "unknown result type\n")
	}
}

func renderStatusText(r contract.StatusResult) string {
	var buf strings.Builder

	buf.WriteString("Environment Overview\n")
	buf.WriteString(strings.Repeat("─", 20) + "\n")
	fmt.Fprintf(&buf, "Projects: %d  |  Sessions: %d (%d active)",
		r.Projects, r.Sessions, r.Lifecycle.Active)

	if r.Health != nil && r.Health.EnvironmentScore > 0 {
		fmt.Fprintf(&buf, "  |  Health: %d/100", r.Health.EnvironmentScore)
	}
	buf.WriteString("\n\n")

	if r.Health == nil {
		return buf.String()
	}

	if len(r.Health.ProjectScores) > 0 {
		buf.WriteString("Lowest Health Projects\n")
		buf.WriteString(strings.Repeat("─", 22) + "\n")
		for i, ps := range r.Health.ProjectScores {
			if i >= 3 {
				break
			}
			fmt.Fprintf(&buf, "  %d. %-15s Score: %d  |  Stale: %.0f%%  |  Activity: %d\n",
				i+1, ps.ProjectName, ps.HealthScore,
				ps.StaleRatio*100, ps.ActivityScore)
		}
		buf.WriteString("\n")
	}

	if len(r.Health.Recommendations) > 0 {
		buf.WriteString("Recommendations:\n")
		for _, rec := range r.Health.Recommendations {
			fmt.Fprintf(&buf, "  - %s\n", rec)
		}
	}

	return buf.String()
}

func renderProjectListText(r contract.ProjectListResult) string {
	if len(r.Projects) == 0 {
		return "No projects found.\n"
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Projects (%d)\n\n", len(r.Projects))

	tw := tabwriter.NewWriter(&buf, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSESSIONS\tACTIVE\tSKILLS\tCOMMANDS\tAGENTS\tSIZE\tLAST ACTIVE\tPATH")
	for _, p := range r.Projects {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\t%d\t%s\t%s\t%s\n",
			p.Name, p.SessionCount, p.ActiveSessionCount,
			p.SkillCount, p.CommandCount, p.AgentCount,
			formatSize(p.TotalSizeBytes), p.LastActiveAt, p.Path)
	}
	_ = tw.Flush()
	return buf.String()
}

func renderProjectDetailText(r contract.ProjectDetailResult) string {
	d := r.ProjectDetail
	var buf strings.Builder
	fmt.Fprintf(&buf, "Project: %s\n", d.Name)
	fmt.Fprintf(&buf, "  Path:       %s\n", d.Path)
	fmt.Fprintf(&buf, "  Claude Dir: %s\n", d.ClaudeRoot)
	fmt.Fprintf(&buf, "  Last Active: %s\n", d.LastActiveAt)
	fmt.Fprintf(&buf, "  Total Size: %s\n", formatSize(d.TotalSizeBytes))
	fmt.Fprintf(&buf, "\n  Sessions:  %d total, %d active\n", d.Sessions.Total, d.Sessions.Active)
	fmt.Fprintf(&buf, "  Resources: %d skills, %d commands, %d agents\n",
		d.Resources.Skills, d.Resources.Commands, d.Resources.Agents)
	return buf.String()
}

func renderSessionListText(r contract.SessionListResult) string {
	if len(r.Sessions) == 0 {
		return "No sessions found.\n"
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Sessions (%d)\n\n", len(r.Sessions))

	tw := tabwriter.NewWriter(&buf, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tPROJECT\tSTATE\tLAST ACTIVE\tSUMMARY")
	for _, s := range r.Sessions {
		summary := truncateText(s.Summary, 50)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			s.ID, s.Project, s.State, s.LastActiveAt, summary)
	}
	_ = tw.Flush()
	return buf.String()
}

func renderSessionDetailText(r contract.SessionDetailResult) string {
	d := r.SessionDetail
	var buf strings.Builder
	fmt.Fprintf(&buf, "Session: %s\n", d.ID)
	fmt.Fprintf(&buf, "  Project: %s\n", d.Project)
	fmt.Fprintf(&buf, "  Path:    %s\n", d.Path)
	if d.Summary != "" {
		fmt.Fprintf(&buf, "  Summary: %s\n", d.Summary)
	}

	fmt.Fprintf(&buf, "\n  Lifecycle\n")
	fmt.Fprintf(&buf, "    State:          %s\n", d.Lifecycle.State)
	fmt.Fprintf(&buf, "    Last Active:    %s\n", d.Lifecycle.LastActiveAt)
	fmt.Fprintf(&buf, "    Active Marker:  %v\n", d.Lifecycle.HasActiveMarker)
	if len(d.Lifecycle.Reasons) > 0 {
		for _, reason := range d.Lifecycle.Reasons {
			fmt.Fprintf(&buf, "    - %s\n", reason)
		}
	}

	fmt.Fprintf(&buf, "\n  Metadata\n")
	fmt.Fprintf(&buf, "    Model:    %s\n", d.Metadata.Model)
	fmt.Fprintf(&buf, "    Version:  %s\n", d.Metadata.Version)
	fmt.Fprintf(&buf, "    Branch:   %s\n", d.Metadata.GitBranch)
	fmt.Fprintf(&buf, "    Created:  %s\n", d.Metadata.CreatedAt)
	fmt.Fprintf(&buf, "    Updated:  %s\n", d.Metadata.UpdatedAt)
	fmt.Fprintf(&buf, "    Duration: %s\n", formatDuration(d.Metadata.DurationSeconds))

	fmt.Fprintf(&buf, "\n  Activity\n")
	fmt.Fprintf(&buf, "    Turns:     %d\n", d.Activity.TurnCount)
	fmt.Fprintf(&buf, "    Messages:  %d\n", d.Activity.UserMessageCount)
	fmt.Fprintf(&buf, "    Tools:     %d\n", d.Activity.ToolCallCount)

	fmt.Fprintf(&buf, "\n  Tokens\n")
	fmt.Fprintf(&buf, "    Input:  %s\n", formatNumber(d.Tokens.Input))
	fmt.Fprintf(&buf, "    Output: %s\n", formatNumber(d.Tokens.Output))
	fmt.Fprintf(&buf, "    Cache:  %s\n", formatNumber(d.Tokens.Cache))
	return buf.String()
}

func renderSkillListText(r contract.SkillListResult) string {
	if len(r.Skills) == 0 {
		return "No skills found.\n"
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Skills (%d)\n\n", len(r.Skills))

	tw := tabwriter.NewWriter(&buf, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tTYPE\tSCOPE\tSTATUS\tPROJECT\tPATH")
	for _, s := range r.Skills {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			s.Name, s.Type, s.Scope, s.Status, s.Project, s.Path)
	}
	_ = tw.Flush()
	return buf.String()
}

func renderAgentListText(r contract.AgentListResult) string {
	if len(r.Agents) == 0 {
		return "No agents found.\n"
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Agents (%d)\n\n", len(r.Agents))

	tw := tabwriter.NewWriter(&buf, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSCOPE\tSTATUS\tPROJECT\tPATH")
	for _, a := range r.Agents {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			a.Name, a.Scope, a.Status, a.Project, a.Path)
	}
	_ = tw.Flush()
	return buf.String()
}

func renderAgentDetailText(r contract.AgentDetailResult) string {
	d := r.AgentDetail
	var buf strings.Builder
	fmt.Fprintf(&buf, "Agent: %s\n", d.Name)
	fmt.Fprintf(&buf, "  Scope:   %s\n", d.Scope)
	fmt.Fprintf(&buf, "  Source:  %s\n", d.Source)
	fmt.Fprintf(&buf, "  Status:  %s\n", d.Status)
	if d.Project != "" {
		fmt.Fprintf(&buf, "  Project: %s\n", d.Project)
	}
	fmt.Fprintf(&buf, "  Path:    %s\n", d.Path)

	fmt.Fprintf(&buf, "\n  Configuration\n")
	fmt.Fprintf(&buf, "    Model:      %s\n", d.Configuration.Model)
	if len(d.Configuration.Tools) > 0 {
		fmt.Fprintf(&buf, "    Tools:      %s\n", strings.Join(d.Configuration.Tools, ", "))
	}
	fmt.Fprintf(&buf, "    Permission: %s\n", d.Configuration.Permission)

	fmt.Fprintf(&buf, "\n  Validation: %v\n", d.Validation.Valid)
	if len(d.Validation.Reasons) > 0 {
		for _, reason := range d.Validation.Reasons {
			fmt.Fprintf(&buf, "    - %s\n", reason)
		}
	}
	return buf.String()
}

func renderThemesText(r contract.ThemesResult) string {
	var buf strings.Builder
	fmt.Fprintln(&buf, "Available themes:")
	fmt.Fprintln(&buf)
	for _, t := range r.Themes {
		marker := "  "
		if t.Current {
			marker = "* "
		}
		fmt.Fprintf(&buf, "  %s%-16s %s\n", marker, t.Name, t.Description)
	}
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "Current: %s\n", r.Current)
	fmt.Fprintln(&buf, "Usage: cc9s --theme <name>  or  CC9S_THEME=<name> cc9s")
	fmt.Fprintln(&buf, "       cc9s --claude-dir <path>  or  CC9S_CLAUDE_DIR=<path> cc9s")
	return buf.String()
}

func renderCleanupText(r contract.CleanupResult) string {
	var buf strings.Builder
	fmt.Fprintln(&buf, "Session Cleanup Preview (dry-run — no data was modified)")
	fmt.Fprintln(&buf)

	fmt.Fprintf(&buf, "  Filters:  state=%s", r.Filters.State)
	if r.Filters.OlderThan != "" {
		fmt.Fprintf(&buf, "  older-than=%s", r.Filters.OlderThan)
	}
	if r.Filters.Project != "" {
		fmt.Fprintf(&buf, "  project=%s", r.Filters.Project)
	}
	fmt.Fprintln(&buf)

	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "Summary")
	fmt.Fprintf(&buf, "  Matched:  %d sessions across %d projects (%s)\n",
		r.Summary.MatchedSessions, r.Summary.MatchedProjects, formatSize(r.Summary.TotalSizeBytes))
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "Recommendations")
	if r.Summary.DeleteCount > 0 {
		fmt.Fprintf(&buf, "  Delete:   %d sessions (safe to remove)\n", r.Summary.DeleteCount)
	}
	if r.Summary.MaybeCount > 0 {
		fmt.Fprintf(&buf, "  Review:   %d sessions (check before deleting)\n", r.Summary.MaybeCount)
	}
	if r.Summary.KeepCount > 0 {
		fmt.Fprintf(&buf, "  Keep:     %d sessions (valuable content)\n", r.Summary.KeepCount)
	}
	if strings.EqualFold(r.Filters.State, "stale") {
		fmt.Fprintln(&buf)
		fmt.Fprintln(&buf, "Tip")
		fmt.Fprintln(&buf, "  Default cleanup targets stale sessions, so this view usually only shows Delete.")
		fmt.Fprintln(&buf, "  Use \"--state completed\" to see Delete / Review / Keep recommendations.")
	}

	if len(r.Projects) > 0 {
		fmt.Fprintln(&buf)
		fmt.Fprintln(&buf, "Projects")
		for _, p := range r.Projects {
			fmt.Fprintf(&buf, "  %-30s %d sessions\n", p.Name, p.SessionCount)
		}
	}

	fmt.Fprintln(&buf)
	printSessionGroup(&buf, r.Sessions, "Delete", "Recommended for deletion")
	printSessionGroup(&buf, r.Sessions, "Maybe", "Review before deleting")
	printSessionGroup(&buf, r.Sessions, "Keep", "Recommended to keep")

	if len(r.Sessions) > 30 {
		fmt.Fprintf(&buf, "  ... showing top 30 of %d sessions (use --json for full list)\n", len(r.Sessions))
	}
	return buf.String()
}

func printSessionGroup(w io.Writer, sessions []contract.CleanupSessionMatch, recommendation string, header string) {
	var group []contract.CleanupSessionMatch
	for _, s := range sessions {
		if s.Recommendation == recommendation {
			group = append(group, s)
		}
	}
	if len(group) == 0 {
		return
	}

	fmt.Fprintf(w, "%s (%d)\n", header, len(group))
	limit := len(group)
	if limit > 10 {
		limit = 10
	}
	for _, s := range group[:limit] {
		reason := ""
		if len(s.Reasons) > 0 {
			reason = s.Reasons[0]
		}
		fmt.Fprintf(w, "  %-12s %-20s %-10s  %s\n", s.ID[:minInt(12, len(s.ID))], s.Project, s.State, reason)
	}
	if len(group) > 10 {
		fmt.Fprintf(w, "  ... and %d more\n", len(group)-10)
	}
	fmt.Fprintln(w)
}

func writeString(w io.Writer, s string) error {
	if w == nil {
		return fmt.Errorf("nil writer")
	}
	_, err := io.WriteString(w, s)
	return err
}
