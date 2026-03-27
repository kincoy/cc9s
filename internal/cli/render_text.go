package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// flushWriter wraps an io.Writer to propagate tabwriter flush errors.
type flushWriter struct {
	w   io.Writer
	err error
}

func (fw *flushWriter) Write(p []byte) (int, error) {
	if fw.err != nil {
		return 0, fw.err
	}
	n, err := fw.w.Write(p)
	if err != nil {
		fw.err = err
	}
	return n, err
}

// renderTextMode dispatches text rendering for each result type.
func renderTextMode(w *os.File, result CommandResult) {
	switch r := result.(type) {
	case StatusResult:
		renderStatusText(w, r)
	case ProjectListResult:
		renderProjectListText(w, r)
	case ProjectDetailResult:
		renderProjectDetailText(w, r)
	case SessionListResult:
		renderSessionListText(w, r)
	case SessionDetailResult:
		renderSessionDetailText(w, r)
	case SkillListResult:
		renderSkillListText(w, r)
	case AgentListResult:
		renderAgentListText(w, r)
	case AgentDetailResult:
		renderAgentDetailText(w, r)
	case ThemesResult:
		renderThemesText(w, r)
	case CleanupResult:
		renderCleanupText(w, r)
	case HelpResult:
		fmt.Fprint(w, r.Text)
	case VersionResult:
		fmt.Fprintf(w, "cc9s v%s\n", r.Version)
	}
}

// --- Status ---

func renderStatusText(w io.Writer, r StatusResult) {
	fmt.Fprintf(w, "Claude Code Environment\n\n")

	fmt.Fprintf(w, "  Projects:   %d\n", r.Projects)
	fmt.Fprintf(w, "  Sessions:   %d\n", r.Sessions)
	fmt.Fprintf(w, "  Resources:  %d\n", r.Resources)
	fmt.Fprintf(w, "  Total Size: %s\n", formatSize(r.TotalSizeBytes))

	fmt.Fprintf(w, "\nLifecycle\n")
	fmt.Fprintf(w, "  Active:    %d\n", r.Lifecycle.Active)
	fmt.Fprintf(w, "  Idle:      %d\n", r.Lifecycle.Idle)
	fmt.Fprintf(w, "  Completed: %d\n", r.Lifecycle.Completed)
	fmt.Fprintf(w, "  Stale:     %d\n", r.Lifecycle.Stale)

	if len(r.Issues) > 0 {
		fmt.Fprintf(w, "\nIssues\n")
		for _, issue := range r.Issues {
			fmt.Fprintf(w, "  ! %s (%d)", issue.Type, issue.Count)
			if issue.Percentage != "" {
				fmt.Fprintf(w, " [%s]", issue.Percentage)
			}
			fmt.Fprintln(w)
			fmt.Fprintf(w, "    %s\n", issue.Suggestion)
		}
	} else {
		fmt.Fprintf(w, "\nNo issues found.\n")
	}

	if len(r.TopProjects) > 0 {
		fmt.Fprintf(w, "\nTop Projects\n")
		for _, tp := range r.TopProjects {
			fmt.Fprintf(w, "  %s  %d sessions (%d active)  %s\n",
				tp.Name, tp.Sessions, tp.Active, formatSize(int64(tp.SizeBytes)))
		}
	}
}

// --- Projects ---

func renderProjectListText(w io.Writer, r ProjectListResult) {
	if len(r.Projects) == 0 {
		fmt.Fprintln(w, "No projects found.")
		return
	}

	fmt.Fprintf(w, "Projects (%d)\n\n", len(r.Projects))

	fw := &flushWriter{w: w}
	tw := tabwriter.NewWriter(fw, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSESSIONS\tACTIVE\tSKILLS\tCOMMANDS\tAGENTS\tSIZE\tLAST ACTIVE\tPATH")
	for _, p := range r.Projects {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\t%d\t%s\t%s\t%s\n",
			p.Name, p.SessionCount, p.ActiveSessionCount,
			p.SkillCount, p.CommandCount, p.AgentCount,
			formatSize(p.TotalSizeBytes), p.LastActiveAt, p.Path)
	}
	tw.Flush()
	if fw.err != nil {
		fmt.Fprintf(os.Stderr, "error: write output: %v\n", fw.err)
	}
}

func renderProjectDetailText(w io.Writer, r ProjectDetailResult) {
	d := r.ProjectDetail
	fmt.Fprintf(w, "Project: %s\n", d.Name)
	fmt.Fprintf(w, "  Path:       %s\n", d.Path)
	fmt.Fprintf(w, "  Claude Dir: %s\n", d.ClaudeRoot)
	fmt.Fprintf(w, "  Last Active: %s\n", d.LastActiveAt)
	fmt.Fprintf(w, "  Total Size: %s\n", formatSize(d.TotalSizeBytes))
	fmt.Fprintf(w, "\n  Sessions:  %d total, %d active\n", d.Sessions.Total, d.Sessions.Active)
	fmt.Fprintf(w, "  Resources: %d skills, %d commands, %d agents\n",
		d.Resources.Skills, d.Resources.Commands, d.Resources.Agents)
}

// --- Sessions ---

func renderSessionListText(w io.Writer, r SessionListResult) {
	if len(r.Sessions) == 0 {
		fmt.Fprintln(w, "No sessions found.")
		return
	}

	fmt.Fprintf(w, "Sessions (%d)\n\n", len(r.Sessions))

	fw := &flushWriter{w: w}
	tw := tabwriter.NewWriter(fw, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tPROJECT\tSTATE\tLAST ACTIVE\tSUMMARY")
	for _, s := range r.Sessions {
		summary := truncateText(s.Summary, 50)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			s.ID, s.Project, s.State, s.LastActiveAt, summary)
	}
	tw.Flush()
	if fw.err != nil {
		fmt.Fprintf(os.Stderr, "error: write output: %v\n", fw.err)
	}
}

func renderSessionDetailText(w io.Writer, r SessionDetailResult) {
	d := r.SessionDetail
	fmt.Fprintf(w, "Session: %s\n", d.ID)
	fmt.Fprintf(w, "  Project: %s\n", d.Project)
	fmt.Fprintf(w, "  Path:    %s\n", d.Path)
	if d.Summary != "" {
		fmt.Fprintf(w, "  Summary: %s\n", d.Summary)
	}

	fmt.Fprintf(w, "\n  Lifecycle\n")
	fmt.Fprintf(w, "    State:          %s\n", d.Lifecycle.State)
	fmt.Fprintf(w, "    Last Active:    %s\n", d.Lifecycle.LastActiveAt)
	fmt.Fprintf(w, "    Active Marker:  %v\n", d.Lifecycle.HasActiveMarker)
	if len(d.Lifecycle.Reasons) > 0 {
		for _, reason := range d.Lifecycle.Reasons {
			fmt.Fprintf(w, "    - %s\n", reason)
		}
	}

	fmt.Fprintf(w, "\n  Metadata\n")
	fmt.Fprintf(w, "    Model:    %s\n", d.Metadata.Model)
	fmt.Fprintf(w, "    Version:  %s\n", d.Metadata.Version)
	fmt.Fprintf(w, "    Branch:   %s\n", d.Metadata.GitBranch)
	fmt.Fprintf(w, "    Created:  %s\n", d.Metadata.CreatedAt)
	fmt.Fprintf(w, "    Updated:  %s\n", d.Metadata.UpdatedAt)
	fmt.Fprintf(w, "    Duration: %s\n", formatDuration(d.Metadata.DurationSeconds))

	fmt.Fprintf(w, "\n  Activity\n")
	fmt.Fprintf(w, "    Turns:     %d\n", d.Activity.TurnCount)
	fmt.Fprintf(w, "    Messages:  %d\n", d.Activity.UserMessageCount)
	fmt.Fprintf(w, "    Tools:     %d\n", d.Activity.ToolCallCount)

	fmt.Fprintf(w, "\n  Tokens\n")
	fmt.Fprintf(w, "    Input:  %s\n", formatNumber(d.Tokens.Input))
	fmt.Fprintf(w, "    Output: %s\n", formatNumber(d.Tokens.Output))
	fmt.Fprintf(w, "    Cache:  %s\n", formatNumber(d.Tokens.Cache))
}

// --- Skills ---

func renderSkillListText(w io.Writer, r SkillListResult) {
	if len(r.Skills) == 0 {
		fmt.Fprintln(w, "No skills found.")
		return
	}

	fmt.Fprintf(w, "Skills (%d)\n\n", len(r.Skills))

	fw := &flushWriter{w: w}
	tw := tabwriter.NewWriter(fw, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tTYPE\tSCOPE\tSTATUS\tPROJECT\tPATH")
	for _, s := range r.Skills {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			s.Name, s.Type, s.Scope, s.Status, s.Project, s.Path)
	}
	tw.Flush()
	if fw.err != nil {
		fmt.Fprintf(os.Stderr, "error: write output: %v\n", fw.err)
	}
}

// --- Agents ---

func renderAgentListText(w io.Writer, r AgentListResult) {
	if len(r.Agents) == 0 {
		fmt.Fprintln(w, "No agents found.")
		return
	}

	fmt.Fprintf(w, "Agents (%d)\n\n", len(r.Agents))

	fw := &flushWriter{w: w}
	tw := tabwriter.NewWriter(fw, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSCOPE\tSTATUS\tPROJECT\tPATH")
	for _, a := range r.Agents {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			a.Name, a.Scope, a.Status, a.Project, a.Path)
	}
	tw.Flush()
	if fw.err != nil {
		fmt.Fprintf(os.Stderr, "error: write output: %v\n", fw.err)
	}
}

func renderAgentDetailText(w io.Writer, r AgentDetailResult) {
	d := r.AgentDetail
	fmt.Fprintf(w, "Agent: %s\n", d.Name)
	fmt.Fprintf(w, "  Scope:   %s\n", d.Scope)
	fmt.Fprintf(w, "  Source:  %s\n", d.Source)
	fmt.Fprintf(w, "  Status:  %s\n", d.Status)
	if d.Project != "" {
		fmt.Fprintf(w, "  Project: %s\n", d.Project)
	}
	fmt.Fprintf(w, "  Path:    %s\n", d.Path)

	fmt.Fprintf(w, "\n  Configuration\n")
	fmt.Fprintf(w, "    Model:      %s\n", d.Configuration.Model)
	if len(d.Configuration.Tools) > 0 {
		fmt.Fprintf(w, "    Tools:      %s\n", strings.Join(d.Configuration.Tools, ", "))
	}
	fmt.Fprintf(w, "    Permission: %s\n", d.Configuration.Permission)

	fmt.Fprintf(w, "\n  Validation: %v\n", d.Validation.Valid)
	if len(d.Validation.Reasons) > 0 {
		for _, reason := range d.Validation.Reasons {
			fmt.Fprintf(w, "    - %s\n", reason)
		}
	}
}

// --- Themes ---

func renderThemesText(w io.Writer, r ThemesResult) {
	fmt.Fprintln(w, "Available themes:")
	fmt.Fprintln(w)
	for _, t := range r.Themes {
		marker := "  "
		if t.Current {
			marker = "* "
		}
		fmt.Fprintf(w, "  %s%-16s %s\n", marker, t.Name, t.Description)
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Current: %s\n", r.Current)
	fmt.Fprintln(w, "Usage: cc9s --theme <name>  or  CC9S_THEME=<name> cc9s")
}

// --- Cleanup ---

func renderCleanupText(w io.Writer, r CleanupResult) {
	fmt.Fprintln(w, "Session Cleanup Preview (dry-run — no data was modified)")
	fmt.Fprintln(w)

	fmt.Fprintf(w, "  Filters:  state=%s", r.Filters.State)
	if r.Filters.OlderThan != "" {
		fmt.Fprintf(w, "  older-than=%s", r.Filters.OlderThan)
	}
	if r.Filters.Project != "" {
		fmt.Fprintf(w, "  project=%s", r.Filters.Project)
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Summary")
	fmt.Fprintf(w, "  Matched:  %d sessions across %d projects (%s)\n",
		r.Summary.MatchedSessions, r.Summary.MatchedProjects, formatSize(r.Summary.TotalSizeBytes))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Recommendations")
	if r.Summary.DeleteCount > 0 {
		fmt.Fprintf(w, "  Delete:   %d sessions (safe to remove)\n", r.Summary.DeleteCount)
	}
	if r.Summary.MaybeCount > 0 {
		fmt.Fprintf(w, "  Review:   %d sessions (check before deleting)\n", r.Summary.MaybeCount)
	}
	if r.Summary.KeepCount > 0 {
		fmt.Fprintf(w, "  Keep:     %d sessions (valuable content)\n", r.Summary.KeepCount)
	}
	if strings.EqualFold(r.Filters.State, "stale") {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Tip")
		fmt.Fprintln(w, "  Default cleanup targets stale sessions, so this view usually only shows Delete.")
		fmt.Fprintln(w, "  Use \"--state completed\" to see Delete / Review / Keep recommendations.")
	}

	if len(r.Projects) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Projects")
		for _, p := range r.Projects {
			fmt.Fprintf(w, "  %-30s %d sessions\n", p.Name, p.SessionCount)
		}
	}

	fmt.Fprintln(w)
	printSessionGroup(w, r.Sessions, "Delete", "Recommended for deletion")
	printSessionGroup(w, r.Sessions, "Maybe", "Review before deleting")
	printSessionGroup(w, r.Sessions, "Keep", "Recommended to keep")

	if len(r.Sessions) > 30 {
		fmt.Fprintf(w, "  ... showing top 30 of %d sessions (use --json for full list)\n", len(r.Sessions))
	}
}

func printSessionGroup(w io.Writer, sessions []CleanupSessionMatch, recommendation string, header string) {
	var group []CleanupSessionMatch
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// --- Formatting helpers ---

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func formatDuration(seconds float64) string {
	if seconds < 1 {
		return "< 1s"
	}
	h := int(seconds / 3600)
	m := int((seconds - float64(h*3600)) / 60)
	s := int(seconds) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func formatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	var result strings.Builder
	length := len(s)
	for i, c := range s {
		if i > 0 && (length-i)%3 == 0 {
			result.WriteByte(',')
		}
		result.WriteRune(c)
	}
	return result.String()
}

func truncateText(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}
