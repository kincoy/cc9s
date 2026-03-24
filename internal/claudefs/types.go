package claudefs

import "time"

// SessionLifecycleState is the user-facing lifecycle state of a session.
type SessionLifecycleState string

const (
	SessionLifecycleActive    SessionLifecycleState = "Active"
	SessionLifecycleIdle      SessionLifecycleState = "Idle"
	SessionLifecycleCompleted SessionLifecycleState = "Completed"
	SessionLifecycleStale     SessionLifecycleState = "Stale"
)

// ActivityWindow defines lifecycle recency thresholds.
type ActivityWindow struct {
	ActiveWindow time.Duration
	IdleWindow   time.Duration
}

// StateEvidenceSummary explains why a lifecycle state was chosen.
type StateEvidenceSummary struct {
	State           SessionLifecycleState
	LastActiveAt    time.Time
	HasActiveMarker bool
	Reasons         []string
}

// SessionLifecycleSnapshot is the shared lifecycle classification for a session.
type SessionLifecycleSnapshot struct {
	State        SessionLifecycleState
	Evidence     StateEvidenceSummary
	ClassifiedAt time.Time
}

// Project represents a Claude Code project.
type Project struct {
	Name               string    // Project name (last path segment, e.g. "cc9s")
	Path               string    // Full decoded path (e.g. "/Users/kinco/go/src/...")
	EncodedPath        string    // Encoded directory name (for filesystem lookup)
	SessionCount       int       // Number of sessions
	ActiveSessionCount int       // Number of active sessions
	LastActiveAt       time.Time // Most recent session active time
	TotalSize          int64     // Total size of all JSONL files (bytes)
	SkillCount         int       // Number of local project skills
	CommandCount       int       // Number of local project commands
	AgentCount         int       // Number of local project agents
	HasSkillsRoot      bool      // Whether .claude/skills exists
	HasCommandsRoot    bool      // Whether .claude/commands exists
	HasAgentsRoot      bool      // Whether .claude/agents exists
}

// Session represents a Claude Code session.
type Session struct {
	ID              string                   // Session ID (JSONL filename without extension)
	ProjectPath     string                   // Full path of the parent project (from JSONL cwd field)
	EncodedPath     string                   // Project encoded path (for filesystem lookup, needed for deletion)
	StartTime       time.Time                // Session start time (file creation time)
	LastActiveAt    time.Time                // Last active time (file modification time)
	EventCount      int                      // Event count (estimated from JSONL line count)
	FileSize        int64                    // JSONL file size
	HasActiveMarker bool                     // Whether a raw active marker exists in ~/.claude/sessions/*.json
	IsActive        bool                     // Whether the session is currently active under lifecycle rules
	Lifecycle       SessionLifecycleSnapshot // Shared lifecycle snapshot for all session-facing surfaces
	Summary         string                   // Session summary (first 80 chars of the first user message)
}

// ScanResult holds the results of a project scan.
type ScanResult struct {
	Projects      []Project     // Project list
	TotalSessions int           // Total session count
	ActiveCount   int           // Active session count (detected via sessions/*.json)
	ScanDuration  time.Duration // Scan duration
	Err           error         // Scan error
}

// GlobalSession is a session with its parent project name embedded.
type GlobalSession struct {
	Session     Session
	ProjectName string
}

// LifecycleSummary counts sessions by lifecycle state.
type LifecycleSummary struct {
	Total     int
	Active    int
	Idle      int
	Completed int
	Stale     int
}

// SessionHealth captures whether a scanned session still looks reliable.
type SessionHealth struct {
	IsReliable bool
	Problem    string
}
