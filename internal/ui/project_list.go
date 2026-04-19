package ui

import (
	"sort"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
)

// SortField sort field enum
type SortField int

const (
	SortByName SortField = iota
	SortByPath
	SortBySessionCount
	SortBySkillCount
	SortByAgentCount
	SortByActivity // default: by most recent activity
	SortByHealth
	SortBySize
)

// ProjectListModel project list view Model
type ProjectListModel struct {
	projects      []claudefs.Project // currently displayed project list (after filtering)
	allProjects   []claudefs.Project // complete project list (filter source)
	filterQuery   string             // current search query
	cursor        int                // currently selected row index
	viewport      viewport.Model     // viewport for scrolling
	loading       bool               // whether data is loading
	sortBy        SortField          // current sort field
	sortAsc       bool               // sort direction
	lastWidth     int                // last rendered width, used to match visible sort columns
	totalSessions int                // total session count
	activeCount   int                // active session count
	showHealthColumn  bool
	showCleanupColumn bool
	projectHealth     map[string]int // projectName -> healthScore
}

// NewProjectListModel creates a new project list Model
func NewProjectListModel() *ProjectListModel {
	return &ProjectListModel{
		loading: true,
		sortBy:  SortByActivity,
		sortAsc: false,
	}
}

func (m *ProjectListModel) Init() tea.Cmd {
	m.viewport = NewViewportWithSize(80, 20) // default size, will be updated in Update(WindowSizeMsg)
	return tea.Batch(
		m.viewport.Init(),
		scanProjectsCmd,
	)
}

// scanProjectsCmd asynchronously scans projects
func scanProjectsCmd() tea.Msg {
	result := claudefs.ScanProjects()
	return projectsLoadedMsg{result: result}
}

func (m *ProjectListModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.SetWidth(msg.Width)
		// Body height = total - header(3) - tabs(2) - footer(1)
		bodyHeight := msg.Height - 6
		if bodyHeight < 1 {
			bodyHeight = 1
		}
		m.viewport.SetHeight(bodyHeight)
		m.updateViewportContent()

	case projectsLoadedMsg:
		m.loading = false
		m.allProjects = append([]claudefs.Project(nil), msg.result.Projects...)
		m.totalSessions = msg.result.TotalSessions
		m.activeCount = msg.result.ActiveCount
		m.sortProjects()
		m.applyFilter()
		if len(m.projects) > 0 {
			m.cursor = 0
		}
		m.updateViewportContent()

	case tea.KeyPressMsg:
		switch msg.String() {
		// Navigation shortcuts
		case "j", "down":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
				m.updateViewportContent()
				EnsureLineVisible(&m.viewport, m.cursor)
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.updateViewportContent()
				EnsureLineVisible(&m.viewport, m.cursor)
			}
		case "G":
			if len(m.projects) > 0 {
				m.cursor = len(m.projects) - 1
				m.updateViewportContent()
				EnsureLineVisible(&m.viewport, m.cursor)
			}
		case "g":
			m.cursor = 0
			m.updateViewportContent()
			EnsureLineVisible(&m.viewport, m.cursor)

		// Half-page and full-page navigation (cursor-based, works with calculateScrollWindow)
		case "ctrl+d":
			halfPage := m.viewport.Height() / 2
			if halfPage < 1 {
				halfPage = 1
			}
			m.cursor += halfPage
			if m.cursor >= len(m.projects) {
				m.cursor = len(m.projects) - 1
			}
			m.updateViewportContent()
			EnsureLineVisible(&m.viewport, m.cursor)
		case "ctrl+u":
			halfPage := m.viewport.Height() / 2
			if halfPage < 1 {
				halfPage = 1
			}
			m.cursor -= halfPage
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.updateViewportContent()
			EnsureLineVisible(&m.viewport, m.cursor)
		case "pgdown":
			page := m.viewport.Height()
			if page < 1 {
				page = 1
			}
			m.cursor += page
			if m.cursor >= len(m.projects) {
				m.cursor = len(m.projects) - 1
			}
			m.updateViewportContent()
			EnsureLineVisible(&m.viewport, m.cursor)
		case "pgup":
			page := m.viewport.Height()
			if page < 1 {
				page = 1
			}
			m.cursor -= page
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.updateViewportContent()
			EnsureLineVisible(&m.viewport, m.cursor)

		// Sort shortcuts
		case "s":
			m.sortBy = nextProjectSortField(m.sortBy, m.lastWidth, m.showHealthColumn)
			m.sortProjects()
			m.applyFilter()
			m.updateViewportContent()
		case "S":
			m.sortAsc = !m.sortAsc
			m.sortProjects()
			m.applyFilter()
			m.updateViewportContent()

		// Enter session list
		case "enter":
			if len(m.projects) > 0 {
				return func() tea.Msg {
					return EnterProjectMsg{Project: m.projects[m.cursor]}
				}
			}
		case "d":
			if len(m.projects) > 0 {
				return func() tea.Msg {
					return ShowProjectDetailMsg{Project: m.projects[m.cursor]}
				}
			}
		}
	}

	return nil
}

// updateViewportContent updates the viewport with rendered project table content
func (m *ProjectListModel) updateViewportContent() {
	if m.loading {
		return
	}

	// Prefer lastWidth (actual rendering width from View()) over viewport width
	width := m.lastWidth
	if width == 0 {
		width = m.viewport.Width()
	}
	if width == 0 {
		width = 80 // default width if not yet sized
	}
	height := m.viewport.Height()
	if height == 0 {
		height = 20 // default height if not yet sized
	}

	content := renderProjectTable(m.projects, m.cursor, width, height, m.sortBy, m.sortAsc, m.showHealthColumn, m.showCleanupColumn, m.projectHealth, m.filterQuery)
	m.viewport.SetContent(content)
}

// View renders the project list view
func (m *ProjectListModel) View(width, height int) string {
	m.lastWidth = width

	// Loading state
	if m.loading {
		return renderLoadingState(width, height)
	}

	return m.viewport.View()
}

// GetStats returns stats info (for AppModel header)
func (m *ProjectListModel) GetStats() (projectCount, totalSessions, activeCount int) {
	return len(m.projects), m.totalSessions, m.activeCount
}

// sortProjects sorts the project list
func (m *ProjectListModel) sortProjects() {
	if len(m.allProjects) == 0 {
		return
	}

	sort.SliceStable(m.allProjects, func(i, j int) bool {
		var less bool
		switch m.sortBy {
		case SortByName:
			less = m.allProjects[i].Name < m.allProjects[j].Name
		case SortByPath:
			less = m.allProjects[i].Path < m.allProjects[j].Path
		case SortBySessionCount:
			less = m.allProjects[i].SessionCount < m.allProjects[j].SessionCount
		case SortBySkillCount:
			less = projectLocalSkillTotal(m.allProjects[i]) < projectLocalSkillTotal(m.allProjects[j])
		case SortByAgentCount:
			less = m.allProjects[i].AgentCount < m.allProjects[j].AgentCount
		case SortByActivity:
			less = m.allProjects[i].LastActiveAt.Before(m.allProjects[j].LastActiveAt)
		case SortByHealth:
			less = m.projectHealth[m.allProjects[i].Name] < m.projectHealth[m.allProjects[j].Name]
		case SortBySize:
			less = m.allProjects[i].TotalSize < m.allProjects[j].TotalSize
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
	m.applyFilter()
	m.updateViewportContent()
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

func (m *ProjectListModel) applyFilter() {
	q := normalizeResourceSearchQuery(m.filterQuery)
	if q == "" {
		m.projects = append([]claudefs.Project(nil), m.allProjects...)
		m.clampCursor()
		return
	}

	var filtered []claudefs.Project
	for _, p := range m.allProjects {
		if strings.Contains(strings.ToLower(p.Name), q) ||
			strings.Contains(strings.ToLower(p.Path), q) ||
			strings.Contains(strings.ToLower(p.EncodedPath), q) {
			filtered = append(filtered, p)
		}
	}
	m.projects = filtered
	m.clampCursor()
}

func nextProjectSortField(current SortField, width int, showHealthColumn bool) SortField {
	order := []SortField{
		SortByName,
		SortBySessionCount,
		SortBySkillCount,
		SortByAgentCount,
		SortByActivity,
		SortBySize,
	}
	if showHealthColumn {
		order = []SortField{
			SortByName,
			SortBySessionCount,
			SortBySkillCount,
			SortByAgentCount,
			SortByActivity,
			SortByHealth,
			SortBySize,
		}
	}
	if width >= 140 {
		order = []SortField{
			SortByName,
			SortByPath,
			SortBySessionCount,
			SortBySkillCount,
			SortByAgentCount,
			SortByActivity,
		}
		if showHealthColumn {
			order = append(order, SortByHealth)
		}
		order = append(order, SortBySize)
	}

	for i, field := range order {
		if field == current {
			return order[(i+1)%len(order)]
		}
	}

	return order[0]
}

func projectLocalSkillTotal(project claudefs.Project) int {
	return project.SkillCount + project.CommandCount
}
