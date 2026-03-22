package claudefs

import "time"

// LogEntry represents a Turn in the log view (a conversation round).
type LogEntry struct {
	TurnNumber   int        // Turn number
	Timestamp    time.Time  // Turn start time
	Duration     int        // Turn duration (milliseconds)
	UserMsg      string     // User message (truncated for display)
	AssistantMsg string     // Assistant reply (truncated for display)
	ToolCalls    []ToolCall // Tool call list
}

// ToolCall represents a single tool invocation.
type ToolCall struct {
	Name    string // Tool name (e.g. Read, Write, Bash)
	Input   string // Parameters (truncated for display)
	Output  string // Result (truncated for display)
	IsError bool   // Whether the call resulted in an error
}
