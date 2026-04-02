package contract

import "github.com/kincoy/cc9s/internal/claudefs"

// OutputMode controls whether a command renders text or JSON.
type OutputMode int

const (
	OutputText OutputMode = iota
	OutputJSON
)

// Result is the interface for all command outputs.
type Result interface {
	isResult()
}

// ErrorPayload is the JSON error response shape.
type ErrorPayload struct {
	Error string `json:"error"`
}

// ProjectListOptions controls projects list filtering/sorting.
type ProjectListOptions struct {
	Limit int
	Sort  string
}

// SessionListOptions controls sessions list filtering/sorting.
type SessionListOptions struct {
	Project string
	State   string
	Limit   int
	Sort    string
}

// SessionInspectOptions identifies a specific session to inspect.
type SessionInspectOptions struct {
	ID string
}

// ProjectInspectOptions identifies a specific project to inspect.
type ProjectInspectOptions struct {
	Target string
}

// SkillListOptions controls skills list filtering.
type SkillListOptions struct {
	Project string
	Scope   string
	Type    string
}

// AgentListOptions controls agents list filtering.
type AgentListOptions struct {
	Project string
	Scope   string
}

// AgentInspectOptions identifies a specific agent to inspect.
type AgentInspectOptions struct {
	Target string
}

// CleanupOptions controls the cleanup preview filters.
type CleanupOptions struct {
	Project   string
	State     string
	OlderThan string
	DryRun    bool
}

// VersionResult is the output of the version command.
type VersionResult struct {
	Version string
}

func (VersionResult) isResult() {}

// StatusResult is the aggregated output of cc9s status.
type StatusResult struct {
	Projects       int                     `json:"projects"`
	Sessions       int                     `json:"sessions"`
	Resources      int                     `json:"resources"`
	TotalSizeBytes int64                   `json:"total_size_bytes"`
	Lifecycle      LifecycleSummary        `json:"lifecycle"`
	Issues         []StatusIssue           `json:"issues"`
	TopProjects    []TopProject            `json:"top_projects"`
	Health         *claudefs.HealthMetrics `json:"health,omitempty"`
}

func (StatusResult) isResult() {}

// LifecycleSummary holds session counts by lifecycle state.
type LifecycleSummary struct {
	Active    int `json:"active"`
	Idle      int `json:"idle"`
	Completed int `json:"completed"`
	Stale     int `json:"stale"`
}

// StatusIssue represents one actionable warning from status.
type StatusIssue struct {
	Type       string `json:"type"`
	Count      int    `json:"count"`
	Percentage string `json:"percentage,omitempty"`
	Suggestion string `json:"suggestion"`
}

// TopProject is one entry in the status top-projects section.
type TopProject struct {
	Name      string `json:"name"`
	Sessions  int    `json:"sessions"`
	Active    int    `json:"active"`
	SizeBytes int64  `json:"size_bytes"`
}

// ProjectListResult is the output of projects list.
type ProjectListResult struct {
	Projects []ProjectListEntry `json:"-"`
}

func (ProjectListResult) isResult() {}

// ProjectListEntry is one row in projects list.
type ProjectListEntry struct {
	Name               string `json:"name"`
	SessionCount       int    `json:"session_count"`
	ActiveSessionCount int    `json:"active_session_count"`
	LastActiveAt       string `json:"last_active_at"`
	SkillCount         int    `json:"skill_count"`
	CommandCount       int    `json:"command_count"`
	AgentCount         int    `json:"agent_count"`
	TotalSizeBytes     int64  `json:"total_size_bytes"`
	Path               string `json:"path"`
}

// ProjectDetailResult is the output of projects inspect.
type ProjectDetailResult struct {
	ProjectDetail ProjectDetail `json:"-"`
}

func (ProjectDetailResult) isResult() {}

// ProjectDetail is the full projection for projects inspect.
type ProjectDetail struct {
	Name           string           `json:"name"`
	Path           string           `json:"path"`
	ClaudeRoot     string           `json:"claude_root"`
	LastActiveAt   string           `json:"last_active_at"`
	TotalSizeBytes int64            `json:"total_size_bytes"`
	Sessions       ProjectSessions  `json:"sessions"`
	Resources      ProjectResources `json:"resources"`
}

// ProjectSessions is the sessions section of a project detail.
type ProjectSessions struct {
	Total  int `json:"total"`
	Active int `json:"active"`
}

// ProjectResources is the resources section of a project detail.
type ProjectResources struct {
	Skills   int `json:"skills"`
	Commands int `json:"commands"`
	Agents   int `json:"agents"`
}

// SessionListResult is the output of sessions list.
type SessionListResult struct {
	Sessions []SessionListEntry `json:"-"`
}

func (SessionListResult) isResult() {}

// SessionListEntry is one row in sessions list.
type SessionListEntry struct {
	ID           string `json:"id"`
	Project      string `json:"project"`
	State        string `json:"state"`
	LastActiveAt string `json:"last_active_at"`
	Summary      string `json:"summary"`
}

// SessionDetailResult is the output of sessions inspect.
type SessionDetailResult struct {
	SessionDetail SessionDetail `json:"-"`
}

func (SessionDetailResult) isResult() {}

// SessionDetail is the full projection for sessions inspect.
type SessionDetail struct {
	ID        string                 `json:"id"`
	Project   string                 `json:"project"`
	Path      string                 `json:"path"`
	Summary   string                 `json:"summary"`
	Lifecycle SessionLifecycleDetail `json:"lifecycle"`
	Metadata  SessionMetadata        `json:"metadata"`
	Activity  SessionActivity        `json:"activity"`
	Tokens    SessionTokens          `json:"tokens"`
}

// SessionLifecycleDetail is the lifecycle section of a session detail.
type SessionLifecycleDetail struct {
	State           string   `json:"state"`
	LastActiveAt    string   `json:"last_active_at"`
	HasActiveMarker bool     `json:"has_active_marker"`
	Reasons         []string `json:"reasons"`
}

// SessionMetadata is the metadata section of a session detail.
type SessionMetadata struct {
	Model           string  `json:"model"`
	Version         string  `json:"version"`
	GitBranch       string  `json:"git_branch"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	DurationSeconds float64 `json:"duration_seconds"`
}

// SessionActivity is the activity section of a session detail.
type SessionActivity struct {
	TurnCount        int            `json:"turn_count"`
	UserMessageCount int            `json:"user_message_count"`
	ToolCallCount    int            `json:"tool_call_count"`
	ToolUsage        map[string]int `json:"tool_usage"`
}

// SessionTokens is the token usage section of a session detail.
type SessionTokens struct {
	Input  int `json:"input"`
	Output int `json:"output"`
	Cache  int `json:"cache"`
}

// SkillListResult is the output of skills list.
type SkillListResult struct {
	Skills []SkillListEntry `json:"-"`
}

func (SkillListResult) isResult() {}

// SkillListEntry is one row in skills list.
type SkillListEntry struct {
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	Scope             string   `json:"scope"`
	Status            string   `json:"status"`
	Project           string   `json:"project"`
	Path              string   `json:"path"`
	ValidationReasons []string `json:"validation_reasons,omitempty"`
}

// AgentListResult is the output of agents list.
type AgentListResult struct {
	Agents []AgentListEntry `json:"-"`
}

func (AgentListResult) isResult() {}

// AgentListEntry is one row in agents list.
type AgentListEntry struct {
	Name              string   `json:"name"`
	Scope             string   `json:"scope"`
	Status            string   `json:"status"`
	Project           string   `json:"project"`
	Path              string   `json:"path"`
	ValidationReasons []string `json:"validation_reasons,omitempty"`
}

// AgentDetailResult is the output of agents inspect.
type AgentDetailResult struct {
	AgentDetail AgentDetail `json:"-"`
}

func (AgentDetailResult) isResult() {}

// AgentDetail is the full projection for agents inspect.
type AgentDetail struct {
	Name          string             `json:"name"`
	Scope         string             `json:"scope"`
	Source        string             `json:"source"`
	Status        string             `json:"status"`
	Project       string             `json:"project"`
	Path          string             `json:"path"`
	Configuration AgentConfiguration `json:"configuration"`
	Validation    AgentValidation    `json:"validation"`
}

// AgentConfiguration is the configuration section of an agent detail.
type AgentConfiguration struct {
	Model      string   `json:"model"`
	Tools      []string `json:"tools"`
	Permission string   `json:"permission"`
}

// AgentValidation is the validation section of an agent detail.
type AgentValidation struct {
	Valid   bool     `json:"valid"`
	Reasons []string `json:"reasons"`
}

// ThemesResult is the output of the themes command.
type ThemesResult struct {
	Themes  []ThemeEntry `json:"-"`
	Current string       `json:"-"`
}

func (ThemesResult) isResult() {}

// ThemeEntry is one row in themes list.
type ThemeEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Current     bool   `json:"current"`
}

// CleanupResult is the output of sessions cleanup --dry-run.
type CleanupResult struct {
	DryRun   bool                  `json:"dry_run"`
	Filters  CleanupFilters        `json:"filters"`
	Summary  CleanupSummary        `json:"summary"`
	Projects []CleanupProjectGroup `json:"projects"`
	Sessions []CleanupSessionMatch `json:"sessions"`
}

func (CleanupResult) isResult() {}

// CleanupFilters describes the filters used for cleanup matching.
type CleanupFilters struct {
	State     string `json:"state"`
	OlderThan string `json:"older_than"`
	Project   string `json:"project"`
}

// CleanupSummary aggregates the cleanup preview totals.
type CleanupSummary struct {
	MatchedSessions int   `json:"matched_sessions"`
	MatchedProjects int   `json:"matched_projects"`
	TotalSizeBytes  int64 `json:"total_size_bytes"`
	DeleteCount     int   `json:"delete_count"`
	MaybeCount      int   `json:"maybe_count"`
	KeepCount       int   `json:"keep_count"`
}

// CleanupProjectGroup is one project group in the cleanup preview.
type CleanupProjectGroup struct {
	Name         string `json:"name"`
	SessionCount int    `json:"session_count"`
}

// CleanupSessionMatch is one matched session in a cleanup preview.
type CleanupSessionMatch struct {
	ID             string   `json:"id"`
	Project        string   `json:"project"`
	State          string   `json:"state"`
	AgeHours       float64  `json:"age_hours"`
	UpdatedAt      string   `json:"updated_at"`
	Recommendation string   `json:"recommendation"`
	Score          int      `json:"score"`
	Reasons        []string `json:"reasons"`
}
