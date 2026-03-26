package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/ui/styles"
	"github.com/kincoy/cc9s/internal/version"
)

// HelpText is the full help output.
const HelpText = `cc9s — Claude Code session manager

Usage:
  cc9s                      Launch TUI (default, no arguments)
  cc9s status               Environment health overview
  cc9s projects list        List all projects
  cc9s projects inspect <name>  Project details (match by name or path)
  cc9s sessions list        List sessions across all projects
  cc9s sessions inspect <id>   Session details (exact ID from list output)
  cc9s sessions cleanup --dry-run  Preview stale/old sessions (read-only)
  cc9s skills list          List skills and commands
  cc9s agents list          List agents
  cc9s agents inspect <name>   Agent details (match by name or path)
  cc9s themes               List available themes
  cc9s version              Print version
  cc9s help                 Print this help

Short flags:
  -h, --help                Show help
  -v, --version             Print version

Commands and flags:
  status                   (no extra flags)
  projects list            --limit <n>  --sort <field>  --json
  projects inspect <name>  --json
  sessions list            --project <name>  --state <state>  --limit <n>  --sort <field>  --json
  sessions inspect <id>    --json
  sessions cleanup         --dry-run  --project <name>  --state <state>  --older-than <dur>  --json
  skills list              --project <name>  --scope <scope>  --type <type>  --json
  agents list              --project <name>  --scope <scope>  --json
  agents inspect <name>    --json

  --json is supported on all commands. Default output is human-readable text.

Enumerations:
  --state <state>          Active, Idle, Completed, Stale (case-insensitive partial match)
  --scope <scope>          User, Project, Plugin (case-insensitive partial match)
  --type <type>            Skill, Command (case-insensitive partial match)
  --sort <field>           projects: name, sessions | sessions: updated, state, project
  --older-than <dur>       Duration, e.g. 72h, 7d, 168h, 30m

Resource aliases:
  projects | project | proj
  sessions | session | ss
  skills   | skill   | sk
  agents   | agent   | ag

Output:
  list commands           -> JSON array of objects
  status / inspect / cleanup -> JSON single object
  errors                  -> {"error":"<message>"}
  All timestamps are RFC 3339. Paths are absolute.

Common patterns:
  cc9s status                              Quick environment health check
  cc9s status --json                        Machine-readable overview
  cc9s sessions list --state active --json  Find active sessions, get full IDs
  cc9s sessions inspect <id> --json         Full session details (model, tokens, lifecycle)
  cc9s sessions cleanup --dry-run           Preview what would be cleaned up
  cc9s projects inspect cc9s               Inspect a specific project
  cc9s skills list --project cc9s --json    Skills for one project
`

// Execute dispatches a parsed command and returns the result.
func Execute(cmd *Command) (CommandResult, error) {
	switch cmd.TopLevel {
	case CmdHelp:
		return HelpResult{Text: HelpText}, nil
	case CmdVersion:
		return VersionResult{Version: version.Version}, nil
	case CmdStatus:
		return executeStatus(cmd)
	case CmdProjects:
		return executeProjects(cmd)
	case CmdSessions:
		return executeSessions(cmd)
	case CmdSkills:
		return executeSkills(cmd)
	case CmdAgents:
		return executeAgents(cmd)
	case CmdThemes:
		return executeThemes(cmd)
	default:
		return nil, CLIError{Message: "unknown command"}
	}
}

// Run is the top-level entry point for CLI execution.
// It returns the exit code (0 for success, 1 for error).
func Run(args []string) int {
	cmd, err := Parse(args)
	if err != nil {
		jsonMode := false
		for _, a := range args {
			if a == "--json" {
				jsonMode = true
				break
			}
		}
		exitWithError(err, jsonMode)
	}
	if cmd == nil {
		// No args — caller should launch TUI
		return 0
	}

	result, err := Execute(cmd)
	if err != nil {
		exitWithError(err, cmd.Output == OutputJSON)
	}

	renderResult(os.Stdout, cmd.Output, result)
	return 0
}

// --- Status ---

func executeStatus(cmd *Command) (CommandResult, error) {
	scanResult := claudefs.ScanProjects()
	if scanResult.Err != nil {
		return nil, CLIError{Message: fmt.Sprintf("scan projects: %v", scanResult.Err)}
	}

	// Scan skills and agents for resource count
	skillResult := claudefs.ScanSkills()
	agentResult := claudefs.ScanAgents()

	totalResources := len(skillResult.Skills) + len(agentResult.Agents)

	// Aggregate lifecycle across all sessions
	allSessions := make([]claudefs.Session, 0)
	for _, project := range scanResult.Projects {
		sessions, err := claudefs.LoadProjectSessions(project.EncodedPath)
		if err != nil {
			continue
		}
		allSessions = append(allSessions, sessions...)
	}

	lifecycle := claudefs.SummarizeLifecycleSessions(allSessions)

	// Calculate total size
	var totalSize int64
	for _, p := range scanResult.Projects {
		totalSize += p.TotalSize
	}

	// Detect issues
	var issues []StatusIssue
	lc := lifecycle
	totalSessions := lc.Active + lc.Idle + lc.Completed + lc.Stale
	if lc.Stale > 0 {
		pct := ""
		if totalSessions > 0 {
			pct = fmt.Sprintf("%.0f%%", float64(lc.Stale)/float64(totalSessions)*100)
		}
		issues = append(issues, StatusIssue{
			Type:       "stale sessions",
			Count:      lc.Stale,
			Percentage: pct,
			Suggestion: "Run: cc9s sessions cleanup --dry-run",
		})
	}

	// Find invalid resources
	invalidSkills := skillResult.InvalidCount
	invalidAgents := agentResult.InvalidCount
	if invalidSkills > 0 {
		issues = append(issues, StatusIssue{
			Type:       "invalid skills",
			Count:      invalidSkills,
			Suggestion: "Run: cc9s skills list --json",
		})
	}
	if invalidAgents > 0 {
		issues = append(issues, StatusIssue{
			Type:       "invalid agents",
			Count:      invalidAgents,
			Suggestion: "Run: cc9s agents list --json",
		})
	}

	// Top projects (by session count, top 5)
	topProjects := make([]TopProject, 0, 5)
	for _, p := range scanResult.Projects {
		topProjects = append(topProjects, TopProject{
			Name:      p.Name,
			Sessions:  p.SessionCount,
			Active:    p.ActiveSessionCount,
			SizeBytes: p.TotalSize,
		})
	}
	sortTopProjects(topProjects)
	if len(topProjects) > 5 {
		topProjects = topProjects[:5]
	}

	return StatusResult{
		Projects:       len(scanResult.Projects),
		Sessions:       totalSessions,
		Resources:      totalResources,
		TotalSizeBytes: totalSize,
		Lifecycle: LifecycleSummary{
			Active:    lc.Active,
			Idle:      lc.Idle,
			Completed: lc.Completed,
			Stale:     lc.Stale,
		},
		Issues:      issues,
		TopProjects: topProjects,
	}, nil
}

// --- Projects ---

func executeProjects(cmd *Command) (CommandResult, error) {
	scanResult := claudefs.ScanProjects()
	if scanResult.Err != nil {
		return nil, CLIError{Message: fmt.Sprintf("scan projects: %v", scanResult.Err)}
	}

	switch cmd.Verb {
	case VerbList:
		return executeProjectsList(cmd, scanResult.Projects)
	case VerbInspect:
		return executeProjectsInspect(cmd, scanResult.Projects)
	default:
		return nil, CLIError{Message: "unknown verb for projects"}
	}
}

func executeProjectsList(cmd *Command, projects []claudefs.Project) (CommandResult, error) {
	entries := make([]ProjectListEntry, 0, len(projects))
	for _, p := range projects {
		entries = append(entries, ProjectListEntry{
			Name:               p.Name,
			SessionCount:       p.SessionCount,
			ActiveSessionCount: p.ActiveSessionCount,
			LastActiveAt:       formatTimeRFC3339(p.LastActiveAt),
			SkillCount:         p.SkillCount,
			CommandCount:       p.CommandCount,
			AgentCount:         p.AgentCount,
			TotalSizeBytes:     p.TotalSize,
			Path:               p.Path,
		})
	}

	if cmd.Sort != "" {
		entries = sortProjectEntries(entries, cmd.Sort)
	}

	if cmd.Limit > 0 && cmd.Limit < len(entries) {
		entries = entries[:cmd.Limit]
	}

	return ProjectListResult{Projects: entries}, nil
}

func executeProjectsInspect(cmd *Command, projects []claudefs.Project) (CommandResult, error) {
	target := strings.ToLower(cmd.Target)

	var found *claudefs.Project
	for i := range projects {
		if strings.ToLower(projects[i].Name) == target || projects[i].Path == cmd.Target {
			found = &projects[i]
			break
		}
	}
	if found == nil {
		return nil, CLIError{Message: fmt.Sprintf("project not found: %s", cmd.Target)}
	}

	// Load full session data for this project
	sessions, err := claudefs.LoadProjectSessions(found.EncodedPath)
	if err != nil {
		return nil, CLIError{Message: fmt.Sprintf("load sessions: %v", err)}
	}

	lifecycle := claudefs.SummarizeLifecycleSessions(sessions)

	return ProjectDetailResult{
		ProjectDetail: ProjectDetail{
			Name:           found.Name,
			Path:           found.Path,
			ClaudeRoot:     found.Path + "/.claude",
			LastActiveAt:   formatTimeRFC3339(found.LastActiveAt),
			TotalSizeBytes: found.TotalSize,
			Sessions: ProjectSessions{
				Total:  found.SessionCount,
				Active: lifecycle.Active,
			},
			Resources: ProjectResources{
				Skills:   found.SkillCount,
				Commands: found.CommandCount,
				Agents:   found.AgentCount,
			},
		},
	}, nil
}

// --- Sessions ---

func executeSessions(cmd *Command) (CommandResult, error) {
	switch cmd.Verb {
	case VerbList:
		return executeSessionsList(cmd)
	case VerbInspect:
		return executeSessionsInspect(cmd)
	case VerbCleanup:
		return executeSessionsCleanup(cmd)
	default:
		return nil, CLIError{Message: "unknown verb for sessions"}
	}
}

func executeSessionsList(cmd *Command) (CommandResult, error) {
	scanResult := claudefs.ScanProjects()
	if scanResult.Err != nil {
		return nil, CLIError{Message: fmt.Sprintf("scan projects: %v", scanResult.Err)}
	}

	var allSessions []SessionListEntry
	for _, project := range scanResult.Projects {
		if cmd.ProjectFilter != "" && !strings.EqualFold(project.Name, cmd.ProjectFilter) {
			continue
		}

		sessions, err := claudefs.LoadProjectSessions(project.EncodedPath)
		if err != nil {
			continue
		}

		for _, s := range sessions {
			if cmd.StateFilter != "" && !claudefs.LifecycleStateMatchesQuery(s.Lifecycle.State, cmd.StateFilter) {
				continue
			}

			summary := s.Summary
			if summary != "" {
				summary = truncateText(summary, 80)
			}

			allSessions = append(allSessions, SessionListEntry{
				ID:           s.ID,
				Project:      project.Name,
				State:        string(s.Lifecycle.State),
				LastActiveAt: formatTimeRFC3339(s.LastActiveAt),
				Summary:      summary,
			})
		}
	}

	if cmd.Sort != "" {
		allSessions = sortSessionEntries(allSessions, cmd.Sort)
	}

	if cmd.Limit > 0 && cmd.Limit < len(allSessions) {
		allSessions = allSessions[:cmd.Limit]
	}

	return SessionListResult{Sessions: allSessions}, nil
}

func executeSessionsInspect(cmd *Command) (CommandResult, error) {
	scanResult := claudefs.ScanProjects()
	if scanResult.Err != nil {
		return nil, CLIError{Message: fmt.Sprintf("scan projects: %v", scanResult.Err)}
	}

	// Find the session across all projects
	var found *claudefs.Session
	var projectName string
	for _, project := range scanResult.Projects {
		sessions, err := claudefs.LoadProjectSessions(project.EncodedPath)
		if err != nil {
			continue
		}
		for i := range sessions {
			if sessions[i].ID == cmd.Target {
				found = &sessions[i]
				projectName = project.Name
				break
			}
		}
		if found != nil {
			break
		}
	}
	if found == nil {
		return nil, CLIError{Message: fmt.Sprintf("session not found: %s", cmd.Target)}
	}

	// Parse full session stats
	stats, err := claudefs.ParseSessionStats(*found)
	if err != nil {
		return nil, CLIError{Message: fmt.Sprintf("parse session stats: %v", err)}
	}

	reasons := make([]string, len(found.Lifecycle.Evidence.Reasons))
	copy(reasons, found.Lifecycle.Evidence.Reasons)

	var durationSeconds float64
	if stats.Duration > 0 {
		durationSeconds = stats.Duration.Seconds()
	}

	return SessionDetailResult{
		SessionDetail: SessionDetail{
			ID:      found.ID,
			Project: projectName,
			Path:    found.ProjectPath,
			Summary: found.Summary,
			Lifecycle: SessionLifecycleDetail{
				State:           string(found.Lifecycle.State),
				LastActiveAt:    formatTimeRFC3339(found.Lifecycle.Evidence.LastActiveAt),
				HasActiveMarker: found.HasActiveMarker,
				Reasons:         reasons,
			},
			Metadata: SessionMetadata{
				Model:           stats.Model,
				Version:         stats.Version,
				GitBranch:       stats.GitBranch,
				CreatedAt:       formatTimeRFC3339(stats.StartTime),
				UpdatedAt:       formatTimeRFC3339(stats.LastActiveTime),
				DurationSeconds: durationSeconds,
			},
			Activity: SessionActivity{
				TurnCount:        stats.TurnCount,
				UserMessageCount: stats.UserMsgCount,
				ToolCallCount:    stats.ToolCallCount,
				ToolUsage:        stats.ToolUsage,
			},
			Tokens: SessionTokens{
				Input:  stats.InputTokens,
				Output: stats.OutputTokens,
				Cache:  stats.CacheTokens,
			},
		},
	}, nil
}

func executeSessionsCleanup(cmd *Command) (CommandResult, error) {
	scanResult := claudefs.ScanProjects()
	if scanResult.Err != nil {
		return nil, CLIError{Message: fmt.Sprintf("scan projects: %v", scanResult.Err)}
	}

	now := time.Now()
	stateFilter := cmd.StateFilter
	if stateFilter == "" {
		stateFilter = "stale" // default cleanup target
	}

	var matches []CleanupSessionMatch
	projectGroups := make(map[string]int)
	var totalSize int64

	for _, project := range scanResult.Projects {
		if cmd.ProjectFilter != "" && !strings.EqualFold(project.Name, cmd.ProjectFilter) {
			continue
		}

		sessions, err := claudefs.LoadProjectSessions(project.EncodedPath)
		if err != nil {
			continue
		}

		projectMatchCount := 0
		for _, s := range sessions {
			if !claudefs.LifecycleStateMatchesQuery(s.Lifecycle.State, stateFilter) {
				continue
			}

			// Check older-than filter (already validated in parse phase)
			if cmd.OlderThan != "" {
				maxAge, _ := parseOlderThanDuration(cmd.OlderThan)
				if now.Sub(s.LastActiveAt) < maxAge {
					continue
				}
			}

			ageHours := now.Sub(s.LastActiveAt).Hours()
			matches = append(matches, CleanupSessionMatch{
				ID:        s.ID,
				Project:   project.Name,
				State:     string(s.Lifecycle.State),
				AgeHours:  ageHours,
				UpdatedAt: formatTimeRFC3339(s.LastActiveAt),
			})
			totalSize += s.FileSize
			projectMatchCount++
		}

		if projectMatchCount > 0 {
			projectGroups[project.Name] = projectMatchCount
		}
	}

	// Build project group list
	projects := make([]CleanupProjectGroup, 0, len(projectGroups))
	for name, count := range projectGroups {
		projects = append(projects, CleanupProjectGroup{
			Name:         name,
			SessionCount: count,
		})
	}

	return CleanupResult{
		DryRun: true,
		Filters: CleanupFilters{
			State:     stateFilter,
			OlderThan: cmd.OlderThan,
			Project:   cmd.ProjectFilter,
		},
		Summary: CleanupSummary{
			MatchedSessions: len(matches),
			MatchedProjects: len(projects),
			TotalSizeBytes:  totalSize,
		},
		Projects: projects,
		Sessions: matches,
	}, nil
}

// --- Skills ---

func executeSkills(cmd *Command) (CommandResult, error) {
	skillResult := claudefs.ScanSkills()
	if skillResult.Err != nil {
		return nil, CLIError{Message: fmt.Sprintf("scan skills: %v", skillResult.Err)}
	}

	switch cmd.Verb {
	case VerbList:
		return executeSkillsList(cmd, skillResult.Skills)
	default:
		return nil, CLIError{Message: "unknown verb for skills (only list is supported in v1)"}
	}
}

func executeSkillsList(cmd *Command, skills []claudefs.SkillResource) (CommandResult, error) {
	entries := make([]SkillListEntry, 0, len(skills))
	for _, s := range skills {
		if cmd.ProjectFilter != "" && !strings.EqualFold(s.ProjectName, cmd.ProjectFilter) {
			continue
		}
		if cmd.ScopeFilter != "" && !scopeMatches(string(s.Scope), cmd.ScopeFilter) {
			continue
		}
		if cmd.TypeFilter != "" && !scopeMatches(string(s.Kind), cmd.TypeFilter) {
			continue
		}

		var validationReasons []string
		if s.Status != claudefs.SkillStatusReady && len(s.ValidationReasons) > 0 {
			validationReasons = s.ValidationReasons
		}

		entries = append(entries, SkillListEntry{
			Name:              s.Name,
			Type:              string(s.Kind),
			Scope:             string(s.Scope),
			Status:            string(s.Status),
			Project:           s.ProjectName,
			Path:              s.Path,
			ValidationReasons: validationReasons,
		})
	}

	if cmd.Limit > 0 && cmd.Limit < len(entries) {
		entries = entries[:cmd.Limit]
	}

	return SkillListResult{Skills: entries}, nil
}

// --- Agents ---

func executeAgents(cmd *Command) (CommandResult, error) {
	agentResult := claudefs.ScanAgents()
	if agentResult.Err != nil {
		return nil, CLIError{Message: fmt.Sprintf("scan agents: %v", agentResult.Err)}
	}

	switch cmd.Verb {
	case VerbList:
		return executeAgentsList(cmd, agentResult.Agents)
	case VerbInspect:
		return executeAgentsInspect(cmd, agentResult.Agents)
	default:
		return nil, CLIError{Message: "unknown verb for agents"}
	}
}

func executeAgentsList(cmd *Command, agents []claudefs.AgentResource) (CommandResult, error) {
	entries := make([]AgentListEntry, 0, len(agents))
	for _, a := range agents {
		if cmd.ProjectFilter != "" && !strings.EqualFold(a.ProjectName, cmd.ProjectFilter) {
			continue
		}
		if cmd.ScopeFilter != "" && !scopeMatches(string(a.Scope), cmd.ScopeFilter) {
			continue
		}

		var validationReasons []string
		if a.Status != claudefs.AgentStatusReady && len(a.ValidationReasons) > 0 {
			validationReasons = a.ValidationReasons
		}

		entries = append(entries, AgentListEntry{
			Name:              a.Name,
			Scope:             string(a.Scope),
			Status:            string(a.Status),
			Project:           a.ProjectName,
			Path:              a.Path,
			ValidationReasons: validationReasons,
		})
	}

	if cmd.Limit > 0 && cmd.Limit < len(entries) {
		entries = entries[:cmd.Limit]
	}

	return AgentListResult{Agents: entries}, nil
}

func executeAgentsInspect(cmd *Command, agents []claudefs.AgentResource) (CommandResult, error) {
	target := strings.ToLower(cmd.Target)

	var found *claudefs.AgentResource
	for i := range agents {
		if strings.ToLower(agents[i].Name) == target || agents[i].Path == cmd.Target {
			found = &agents[i]
			break
		}
	}
	if found == nil {
		return nil, CLIError{Message: fmt.Sprintf("agent not found: %s", cmd.Target)}
	}

	return AgentDetailResult{
		AgentDetail: AgentDetail{
			Name:    found.Name,
			Scope:   string(found.Scope),
			Source:  string(found.Source),
			Status:  string(found.Status),
			Project: found.ProjectName,
			Path:    found.Path,
			Configuration: AgentConfiguration{
				Model:      found.Model,
				Tools:      found.Tools,
				Permission: found.PermissionMode,
			},
			Validation: AgentValidation{
				Valid:   found.Status == claudefs.AgentStatusReady,
				Reasons: found.ValidationReasons,
			},
		},
	}, nil
}

// --- Themes ---

func executeThemes(cmd *Command) (CommandResult, error) {
	names := styles.AvailableThemes()
	descriptions := themeDescriptions()
	entries := make([]ThemeEntry, 0, len(names))
	for _, name := range names {
		entries = append(entries, ThemeEntry{
			Name:        name,
			Description: descriptions[name],
			Current:     name == styles.CurrentThemeName,
		})
	}
	return ThemesResult{
		Themes:  entries,
		Current: styles.CurrentThemeName,
	}, nil
}

func themeDescriptions() map[string]string {
	return map[string]string{
		"default":       "Optimized default, transparent-friendly",
		"dark-solid":    "Forced dark background",
		"high-contrast": "Maximum readability",
		"gruvbox":       "Warm retro palette",
	}
}

// --- Helpers ---

func formatTimeRFC3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// scopeMatches checks case-insensitive partial match (same pattern as LifecycleStateMatchesQuery).
func scopeMatches(value, query string) bool {
	v := strings.ToLower(value)
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" || v == "" {
		return false
	}
	return strings.Contains(v, q) || strings.Contains(q, v)
}

// sortProjectEntries sorts project entries by the requested field.
func sortProjectEntries(entries []ProjectListEntry, field string) []ProjectListEntry {
	sorted := make([]ProjectListEntry, len(entries))
	copy(sorted, entries)
	switch field {
	case "name":
		sort.Slice(sorted, func(i, j int) bool {
			return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
		})
	case "sessions":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].SessionCount > sorted[j].SessionCount
		})
	}
	return sorted
}

func sortTopProjects(entries []TopProject) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Sessions != entries[j].Sessions {
			return entries[i].Sessions > entries[j].Sessions
		}
		if entries[i].Active != entries[j].Active {
			return entries[i].Active > entries[j].Active
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
}

// sortSessionEntries sorts session entries by the requested field.
func sortSessionEntries(entries []SessionListEntry, field string) []SessionListEntry {
	sorted := make([]SessionListEntry, len(entries))
	copy(sorted, entries)
	switch field {
	case "updated":
		sort.Slice(sorted, func(i, j int) bool {
			ti, ei := time.Parse(time.RFC3339, sorted[i].LastActiveAt)
			tj, ej := time.Parse(time.RFC3339, sorted[j].LastActiveAt)
			if ei != nil && ej != nil {
				return false
			}
			if ei != nil {
				return false
			}
			if ej != nil {
				return true
			}
			return ti.After(tj)
		})
	case "state":
		sort.Slice(sorted, func(i, j int) bool {
			return strings.ToLower(sorted[i].State) < strings.ToLower(sorted[j].State)
		})
	case "project":
		sort.Slice(sorted, func(i, j int) bool {
			return strings.ToLower(sorted[i].Project) < strings.ToLower(sorted[j].Project)
		})
	}
	return sorted
}
