package ui

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
)

// SessionSortField session sort field enum
type SessionSortField int

const (
	SortBySessionActivity SessionSortField = iota
	SortBySessionID
	SortByEventCount
	SortBySessionSize
)

// SessionListModel session list view Model (unified, supports context filtering)
type SessionListModel struct {
	context      Context              // current filter context
	sessions     []claudefs.GlobalSession // currently displayed session list (after context + search filtering)
	allSessions  []claudefs.GlobalSession // complete session list (load source)
	contextSessions []claudefs.GlobalSession // context-filtered results (search filter source)
	filterQuery  string               // current search query
	cursor       int
	selectedRows map[int]struct{}
	loading      bool
	sortBy       SessionSortField
	sortAsc      bool
}

// sessionsLoadedMsg session data loaded message
type sessionsLoadedMsg struct {
	sessions []claudefs.GlobalSession
	err      error
}

// NewSessionListModel creates a new session list Model (default context=all)
func NewSessionListModel() *SessionListModel {
	return &SessionListModel{
		context: Context{Type: ContextAll},
		loading: true,
		sortBy:  SortBySessionActivity,
	}
}

// NewSessionListModelForProject creates a Model with a specific project context
func NewSessionListModelForProject(projectName string) *SessionListModel {
	return &SessionListModel{
		context: Context{Type: ContextProject, Value: projectName},
		loading: true,
		sortBy:  SortBySessionActivity,
	}
}

// GetContext returns the current context
func (m *SessionListModel) GetContext() Context {
	return m.context
}

// SetContext sets context and re-filters
func (m *SessionListModel) SetContext(ctx Context) tea.Cmd {
	m.context = ctx
	m.filterQuery = ""
	m.selectedRows = nil
	m.applyContext()
	if len(m.allSessions) == 0 {
		return loadAllSessionsCmd()
	}
	return nil
}

// ShowProjectColumn whether to show the PROJECT column
func (m *SessionListModel) ShowProjectColumn() bool {
	return m.context.Type == ContextAll
}

func (m *SessionListModel) Init() tea.Cmd {
	return loadAllSessionsCmd()
}

func (m *SessionListModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case sessionsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.sessions = nil
			m.allSessions = nil
			m.contextSessions = nil
		} else {
			m.allSessions = msg.sessions
			m.sortGlobalSessions(m.allSessions)
			m.applyContext()
			if len(m.sessions) > 0 {
				m.cursor = 0
			}
		}

	case tea.KeyPressMsg:
		switch msg.String() {
		// Navigation shortcuts
		case "j", "down":
			if m.cursor < len(m.sessions)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "G":
			if len(m.sessions) > 0 {
				m.cursor = len(m.sessions) - 1
			}
		case "g":
			m.cursor = 0

		// Sort shortcuts
		case "s":
			m.sortBy = (m.sortBy + 1) % 4
			m.sortGlobalSessions(m.allSessions)
			m.applyContext()
		case "S":
			m.sortAsc = !m.sortAsc
			m.sortGlobalSessions(m.allSessions)
			m.applyContext()

		// Enter session
		case "enter":
			if len(m.sessions) > 0 {
				gs := m.sessions[m.cursor]

				if gs.Session.IsActive {
					projectName := gs.ProjectName
					if projectName == "" {
						projectName = "unknown"
					}
					dialog := NewConfirmDialogModel(
						"⚠ Session Already Active",
						[]string{
							fmt.Sprintf("Session: %s", claudefs.FormatSessionID(gs.Session.ID, 30)),
							fmt.Sprintf("Project: %s", projectName),
							"",
							"This session is currently running.",
							"Entering may cause conflicts.",
							"",
							"Continue anyway?",
						},
						resumeSessionCmd(gs.Session),
						closeDialogCmd(),
					)

					return func() tea.Msg {
						return ShowConfirmDialogMsg{Dialog: dialog}
					}
				}

				return resumeSessionCmd(gs.Session)
			}

		// View details
		case "d":
			if len(m.sessions) > 0 {
				session := m.sessions[m.cursor].Session
				return func() tea.Msg {
					return ShowDetailMsg{Session: session}
				}
			}

		// View log
		case "l":
			if len(m.sessions) > 0 {
				session := m.sessions[m.cursor].Session
				return func() tea.Msg {
					return ShowLogMsg{Session: session}
				}
			}

		// Return to project list
		case "esc":
			return func() tea.Msg {
				return BackToProjectsMsg{}
			}

		// Multi-select
		case "space":
			if len(m.sessions) > 0 {
				m.ToggleSelect(m.cursor)
			}

		// Delete
		case "ctrl+d":
			return m.deleteSelectedCmd()
		}
	}

	return nil
}

// View renders the session list view
func (m *SessionListModel) View(width, height int) string {
	if m.loading {
		return renderCenteredText("Loading sessions...", width, height)
	}

	if len(m.sessions) == 0 {
		if m.context.Type == ContextProject {
			return renderCenteredText(
				fmt.Sprintf("No sessions found in project: %s", m.context.Value),
				width, height,
			)
		}
		return renderCenteredText("No sessions found", width, height)
	}

	// Get context label
	contextLabel := ""
	if m.context.Type == ContextAll {
		contextLabel = "All Projects"
	} else if m.context.Type == ContextProject {
		contextLabel = m.context.Value
	}

	return renderSessionTable(m.sessions, m.cursor, width, height, m.selectedRows, m.ShowProjectColumn(), m.sortBy, m.sortAsc, contextLabel)
}

// Reload reloads all session data
func (m *SessionListModel) Reload() tea.Cmd {
	return loadAllSessionsCmd()
}

// applyContext filters allSessions by context -> contextSessions -> sessions
func (m *SessionListModel) applyContext() {
	if m.context.Type == ContextAll {
		m.contextSessions = m.allSessions
	} else {
		var filtered []claudefs.GlobalSession
		for _, gs := range m.allSessions {
			if gs.ProjectName == m.context.Value {
				filtered = append(filtered, gs)
			}
		}
		m.contextSessions = filtered
	}
	// Apply search filter on context-filtered results
	m.applySearchFilter()
}

// applySearchFilter applies search filter on contextSessions -> sessions
func (m *SessionListModel) applySearchFilter() {
	if m.filterQuery == "" {
		m.sessions = m.contextSessions
		m.clampCursor()
		return
	}

	q := strings.ToLower(m.filterQuery)
	var filtered []claudefs.GlobalSession
	for _, gs := range m.contextSessions {
		if m.matchesFilter(gs, q) {
			filtered = append(filtered, gs)
		}
	}
	m.sessions = filtered
	m.clampCursor()
}

// ApplyFilter sets search query and re-filters
func (m *SessionListModel) ApplyFilter(query string) {
	oldQuery := m.filterQuery
	m.filterQuery = query
	if query != oldQuery {
		m.selectedRows = nil
	}
	m.applySearchFilter()
}

// matchesFilter checks if a GlobalSession matches the search query
func (m *SessionListModel) matchesFilter(gs claudefs.GlobalSession, q string) bool {
	s := gs.Session
	if strings.Contains(strings.ToLower(s.ID), q) {
		return true
	}
	if strings.Contains(strings.ToLower(gs.ProjectName), q) {
		return true
	}
	if strings.Contains(strings.ToLower(s.ProjectPath), q) {
		return true
	}
	if s.IsActive && strings.Contains("active", q) {
		return true
	}
	if !s.IsActive && strings.Contains("completed", q) {
		return true
	}
	return false
}

// clampCursor ensures cursor is within valid range
func (m *SessionListModel) clampCursor() {
	if len(m.sessions) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.sessions) {
		m.cursor = len(m.sessions) - 1
	}
}

// GetFilterStats returns filter stats (filtered/total)
func (m *SessionListModel) GetFilterStats() (filtered, total int) {
	return len(m.sessions), len(m.contextSessions)
}

// ToggleSelect toggles row selection
func (m *SessionListModel) ToggleSelect(idx int) {
	if m.selectedRows == nil {
		m.selectedRows = make(map[int]struct{})
	}
	if _, ok := m.selectedRows[idx]; ok {
		delete(m.selectedRows, idx)
	} else {
		m.selectedRows[idx] = struct{}{}
	}
}

// GetSelectedSessions returns selected sessions (returns cursor row if none selected)
func (m *SessionListModel) GetSelectedSessions() []claudefs.GlobalSession {
	if len(m.selectedRows) == 0 {
		if len(m.sessions) > 0 {
			return []claudefs.GlobalSession{m.sessions[m.cursor]}
		}
		return nil
	}

	var result []claudefs.GlobalSession
	for idx := range m.selectedRows {
		if idx < len(m.sessions) {
			result = append(result, m.sessions[idx])
		}
	}
	return result
}

// ClearSelection clears all selections
func (m *SessionListModel) ClearSelection() {
	m.selectedRows = nil
}

// HasSelection whether any row is selected
func (m *SessionListModel) HasSelection() bool {
	return len(m.selectedRows) > 0
}

// SelectedCount returns the number of selected rows
func (m *SessionListModel) SelectedCount() int {
	return len(m.selectedRows)
}

// sortGlobalSessions sorts a GlobalSession list
func (m *SessionListModel) sortGlobalSessions(sessions []claudefs.GlobalSession) {
	if len(sessions) == 0 {
		return
	}

	sort.SliceStable(sessions, func(i, j int) bool {
		var less bool
		switch m.sortBy {
		case SortBySessionActivity:
			less = sessions[i].Session.LastActiveAt.Before(sessions[j].Session.LastActiveAt)
		case SortBySessionID:
			less = sessions[i].Session.ID < sessions[j].Session.ID
		case SortByEventCount:
			less = sessions[i].Session.EventCount < sessions[j].Session.EventCount
		case SortBySessionSize:
			less = sessions[i].Session.FileSize < sessions[j].Session.FileSize
		}
		if m.sortAsc {
			return less
		}
		return !less
	})
}

// deleteSelectedCmd builds the delete command (with active session protection)
func (m *SessionListModel) deleteSelectedCmd() tea.Cmd {
	sessions := m.GetSelectedSessions()
	if len(sessions) == 0 {
		return nil
	}

	// Check for active sessions
	var activeNames []string
	for _, gs := range sessions {
		if gs.Session.IsActive {
			name := gs.ProjectName + "/" + claudefs.FormatSessionID(gs.Session.ID, 8)
			activeNames = append(activeNames, name)
		}
	}

	if len(activeNames) > 0 {
		dialog := NewAlertDialogModel(
			"Cannot Delete Active Sessions",
			[]string{
				"The following sessions are currently running:",
				fmt.Sprintf("  %s", strings.Join(activeNames, ", ")),
				"",
				"Please exit them before deleting.",
			},
		)
		return func() tea.Msg { return ShowConfirmDialogMsg{Dialog: dialog} }
	}

	// Build delete targets
	var targets []claudefs.DeleteTarget
	for _, gs := range sessions {
		targets = append(targets, claudefs.DeleteTarget{
			SessionID:   gs.Session.ID,
			EncodedPath: gs.Session.EncodedPath,
			IsActive:    gs.Session.IsActive,
		})
	}

	count := len(targets)
	label := "session"
	if count > 1 {
		label = "sessions"
	}

	dialog := NewConfirmDialogModel(
		fmt.Sprintf("Delete %d %s?", count, label),
		[]string{
			fmt.Sprintf("This will permanently delete %d %s.", count, label),
			"This action cannot be undone.",
		},
		func() tea.Msg { return DeleteSessionsMsg{Targets: targets} },
		closeDialogCmd(),
	)
	return func() tea.Msg { return ShowConfirmDialogMsg{Dialog: dialog} }
}

// loadAllSessionsCmd asynchronously loads sessions from all projects
func loadAllSessionsCmd() tea.Cmd {
	return func() tea.Msg {
	result := claudefs.ScanProjects()
	if result.Err != nil {
		return sessionsLoadedMsg{err: result.Err}
	}

	var sessions []claudefs.GlobalSession
	for _, proj := range result.Projects {
		projSessions, err := claudefs.LoadProjectSessions(proj.EncodedPath)
		if err != nil {
			log.Printf("skip project %s: %v", proj.Name, err)
			continue
		}
		for _, s := range projSessions {
			sessions = append(sessions, claudefs.GlobalSession{
				Session:     s,
				ProjectName: proj.Name,
			})
		}
	}
	return sessionsLoadedMsg{sessions: sessions}
	}
}

// resumeSessionCmd resumes a Claude Code session
func resumeSessionCmd(session claudefs.Session) tea.Cmd {
	return func() tea.Msg {
		clearScreen()

		cmd := exec.Command("claude", "--resume", session.ID)
		cmd.Dir = session.ProjectPath

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		result := tea.ExecProcess(cmd, func(err error) tea.Msg {
			clearScreen()

			var fullErr error = err
			if err != nil && stderr.Len() > 0 {
				stderrMsg := strings.TrimSpace(stderr.String())
				if stderrMsg != "" {
					fullErr = fmt.Errorf("%s: %w", stderrMsg, err)
				}
			}

			return sessionResumedMsg{
				sessionID: session.ID,
				err:       fullErr,
			}
		})()

		return result
	}
}

// closeDialogCmd closes the dialog
func closeDialogCmd() tea.Cmd {
	return func() tea.Msg {
		return CloseDialogMsg{}
	}
}

// clearScreen clears the terminal screen
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
