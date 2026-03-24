package ui

import "github.com/kincoy/cc9s/internal/claudefs"

// ContextType context type enum
type ContextType int

const (
	ContextAll     ContextType = iota // Show sessions from all projects
	ContextProject                    // Show sessions from a specific project
)

// Context session filter context
type Context struct {
	Type  ContextType
	Value string // Project name when ContextProject, empty when ContextAll
}

// Cross-Model message definitions

// EnterProjectMsg enters the session list of a project
type EnterProjectMsg struct {
	Project claudefs.Project
}

// BackToProjectsMsg returns to the project list
type BackToProjectsMsg struct{}

// projectsLoadedMsg project data loaded
type projectsLoadedMsg struct {
	result claudefs.ScanResult
}

// ShowConfirmDialogMsg shows the confirm dialog
type ShowConfirmDialogMsg struct {
	Dialog *ConfirmDialogModel
}

// CloseDialogMsg closes the dialog
type CloseDialogMsg struct{}

// sessionResumedMsg message after Claude Code exits
type sessionResumedMsg struct {
	sessionID string
	err       error
}

// ShowDetailMsg shows session detail panel
type ShowDetailMsg struct {
	Session claudefs.Session
}

// CloseDetailMsg closes the detail panel
type CloseDetailMsg struct{}

// ShowSkillDetailMsg shows skill detail panel.
type ShowSkillDetailMsg struct {
	Skill claudefs.SkillResource
}

// CloseSkillDetailMsg closes the skill detail panel.
type CloseSkillDetailMsg struct{}

// EditSkillMsg opens the selected skill in an external editor.
type EditSkillMsg struct {
	Skill claudefs.SkillResource
}

// SkillEditorFinishedMsg is emitted after the editor exits.
type SkillEditorFinishedMsg struct {
	Skill claudefs.SkillResource
	Err   error
}

// statsLoadedMsg session stats loaded
type statsLoadedMsg struct {
	stats *claudefs.SessionStats
	err   error
}

// ShowLogMsg shows session log view
type ShowLogMsg struct {
	Session claudefs.Session
}

// CloseLogMsg closes the log view
type CloseLogMsg struct{}

// logLoadedMsg session log data loaded
type logLoadedMsg struct {
	entries []claudefs.LogEntry
	total   int
	err     error
}

// SwitchResourceMsg switches resource type
type SwitchResourceMsg struct {
	Resource ResourceType
}

// SwitchContextMsg switches Context
type SwitchContextMsg struct {
	Context
}

// DeleteSessionsMsg delete sessions request
type DeleteSessionsMsg struct {
	Targets []claudefs.DeleteTarget
}

// SessionsDeletedMsg sessions deleted
type SessionsDeletedMsg struct {
	Deleted int
	Errs    []error
}
