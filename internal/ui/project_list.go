package ui

import (
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
)

// SortField sort field enum
type SortField int

const (
	SortByActivity SortField = iota // default: by most recent activity
	SortByName
	SortBySessionCount
	SortBySize
)

// ProjectListModel project list view Model
type ProjectListModel struct {
	projects      []claudefs.Project // currently displayed project list (after filtering)
	allProjects   []claudefs.Project // complete project list (filter source)
	filterQuery   string             // current search query
	cursor        int                // currently selected row index
	loading       bool               // whether data is loading
	sortBy        SortField          // current sort field
	sortAsc       bool               // sort direction
	totalSessions int                // total session count
	activeCount   int                // active session count
}

// NewProjectListModel creates a new project list Model
func NewProjectListModel() *ProjectListModel {
	return &ProjectListModel{
		loading: true,
		sortBy:  SortByActivity,
	}
}

func (m *ProjectListModel) Init() tea.Cmd {
	return scanProjectsCmd
}

// scanProjectsCmd asynchronously scans projects
func scanProjectsCmd() tea.Msg {
	result := claudefs.ScanProjects()
	return projectsLoadedMsg{result: result}
}

func (m *ProjectListModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case projectsLoadedMsg:
		m.loading = false
		m.projects = msg.result.Projects
		m.allProjects = msg.result.Projects
		m.totalSessions = msg.result.TotalSessions
		m.activeCount = msg.result.ActiveCount
		if len(m.projects) > 0 {
			m.cursor = 0
		}

	case tea.KeyPressMsg:
		switch msg.String() {
		// Navigation shortcuts
		case "j", "down":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "G":
			if len(m.projects) > 0 {
				m.cursor = len(m.projects) - 1
			}
		case "g":
			m.cursor = 0

		// Sort shortcuts
		case "s":
			m.sortBy = (m.sortBy + 1) % 4
			m.sortProjects()
		case "S":
			m.sortAsc = !m.sortAsc
			m.sortProjects()

		// Enter session list
		case "enter":
			if len(m.projects) > 0 {
				return func() tea.Msg {
					return EnterProjectMsg{Project: m.projects[m.cursor]}
				}
			}
		}
	}

	return nil
}

// View renders the project list view
func (m *ProjectListModel) View(width, height int) string {
	// Loading state
	if m.loading {
		return renderLoadingState(width, height)
	}

	return renderProjectTable(m.projects, m.cursor, width, height, m.sortBy, m.sortAsc)
}

// GetStats returns stats info (for AppModel header)
func (m *ProjectListModel) GetStats() (projectCount, totalSessions, activeCount int) {
	return len(m.projects), m.totalSessions, m.activeCount
}

// sortProjects sorts the project list
func (m *ProjectListModel) sortProjects() {
	if len(m.projects) == 0 {
		return
	}

	sort.SliceStable(m.projects, func(i, j int) bool {
		var less bool
		switch m.sortBy {
		case SortByActivity:
			less = m.projects[i].LastActiveAt.Before(m.projects[j].LastActiveAt)
		case SortByName:
			less = m.projects[i].Name < m.projects[j].Name
		case SortBySessionCount:
			less = m.projects[i].SessionCount < m.projects[j].SessionCount
		case SortBySize:
			less = m.projects[i].TotalSize < m.projects[j].TotalSize
		}
		if m.sortAsc {
			return less
		}
		return !less
	})
}

// renderLoadingState renders the loading state
func renderLoadingState(width, height int) string {
	return renderCenteredText("Loading projects...", width, height)
}

// ApplyFilter filters the project list by query
func (m *ProjectListModel) ApplyFilter(query string) {
	m.filterQuery = query
	q := normalizeResourceSearchQuery(query)
	if q == "" {
		m.projects = m.allProjects
		m.clampCursor()
		return
	}

	var filtered []claudefs.Project
	for _, p := range m.allProjects {
		if strings.Contains(strings.ToLower(p.Name), q) ||
			strings.Contains(strings.ToLower(p.Path), q) {
			filtered = append(filtered, p)
		}
	}
	m.projects = filtered
	m.clampCursor()
}

// clampCursor ensures cursor is within valid range
func (m *ProjectListModel) clampCursor() {
	if len(m.projects) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.projects) {
		m.cursor = len(m.projects) - 1
	}
}

// GetFilterStats returns filter stats (filtered/total)
func (m *ProjectListModel) GetFilterStats() (filtered, total int) {
	return len(m.projects), len(m.allProjects)
}

func (m *ProjectListModel) HasActiveFilter() bool {
	return strings.TrimSpace(normalizeResourceSearchQuery(m.filterQuery)) != ""
}
