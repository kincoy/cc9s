package claudefs

import "time"

// Project represents a Claude Code project.
type Project struct {
	Name         string    // Project name (last path segment, e.g. "cc9s")
	Path         string    // Full decoded path (e.g. "/Users/kinco/go/src/...")
	EncodedPath  string    // Encoded directory name (for filesystem lookup)
	SessionCount int       // Number of sessions
	LastActiveAt time.Time // Most recent session active time
	TotalSize    int64     // Total size of all JSONL files (bytes)
}

// Session represents a Claude Code session.
type Session struct {
	ID           string    // Session ID (JSONL filename without extension)
	ProjectPath  string    // Full path of the parent project (from JSONL cwd field)
	EncodedPath  string    // Project encoded path (for filesystem lookup, needed for deletion)
	StartTime    time.Time // Session start time (file creation time)
	LastActiveAt time.Time // Last active time (file modification time)
	EventCount   int       // Event count (estimated from JSONL line count)
	FileSize     int64     // JSONL file size
	IsActive     bool      // Whether the session is active (sessions/*.json exists)
	Summary      string    // Session summary (first 80 chars of the first user message)
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
