package ui

import (
	"sort"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// AgentSortField agent sort field enum.
type AgentSortField int

const (
	SortByAgentName AgentSortField = iota
	SortByAgentStatus
	SortByAgentScope
)

// AgentListModel agent list view model.
type AgentListModel struct {
	state     DefaultResourceListState[claudefs.AgentResource]
	loadErr   error
	sortBy    AgentSortField
	sortAsc   bool
	viewport  viewport.Model
	lastWidth int
}

type agentsLoadedMsg struct {
	result claudefs.AgentScanResult
}

func NewAgentListModel() *AgentListModel {
	return &AgentListModel{
		state:   NewDefaultResourceListState[claudefs.AgentResource](),
		sortBy:  SortByAgentName,
		sortAsc: true,
	}
}

func (m *AgentListModel) Init() tea.Cmd {
	m.viewport = NewViewportWithSize(80, 20) // default size, updated on WindowSizeMsg
	return tea.Batch(
		m.viewport.Init(),
		scanAgentsCmd,
	)
}

func scanAgentsCmd() tea.Msg {
	return agentsLoadedMsg{result: claudefs.ScanAgents()}
}

func (m *AgentListModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.SetWidth(msg.Width)
		// Body height = total - header (3) - footer (1)
		// Body height = total - header(3) - tabs(2) - footer(1)
		bodyHeight := msg.Height - 6
		if bodyHeight < 1 {
			bodyHeight = 1
		}
		m.viewport.SetHeight(bodyHeight)
		m.updateViewportContent()

	case agentsLoadedMsg:
		m.loadErr = msg.result.Err
		items := append([]claudefs.AgentResource(nil), msg.result.Agents...)
		m.sortAgents(items)
		m.state.SetItems(items, m.agentHooks())
		m.updateViewportContent()
		return func() tea.Msg { return StopLoadingMsg{Resource: ResourceAgents} }

	case tea.KeyPressMsg:
		if m.loadErr != nil {
			return nil
		}

		switch msg.String() {
		case "j", "down":
			if m.state.Cursor < len(m.state.VisibleItems)-1 {
				m.state.Cursor++
				m.updateViewportContent()
				EnsureLineVisible(&m.viewport, m.state.Cursor)
			}
		case "k", "up":
			if m.state.Cursor > 0 {
				m.state.Cursor--
				m.updateViewportContent()
				EnsureLineVisible(&m.viewport, m.state.Cursor)
			}
		case "G":
			if len(m.state.VisibleItems) > 0 {
				m.state.Cursor = len(m.state.VisibleItems) - 1
				m.updateViewportContent()
				EnsureLineVisible(&m.viewport, m.state.Cursor)
			}
		case "g":
			m.state.Cursor = 0
			m.updateViewportContent()
			EnsureLineVisible(&m.viewport, m.state.Cursor)

		// Half-page and full-page navigation
		case "ctrl+d":
			halfPage := m.viewport.Height() / 2
			if halfPage < 1 {
				halfPage = 1
			}
			m.state.Cursor += halfPage
			if m.state.Cursor >= len(m.state.VisibleItems) {
				m.state.Cursor = len(m.state.VisibleItems) - 1
			}
			if m.state.Cursor < 0 {
				m.state.Cursor = 0
			}
			m.updateViewportContent()
			EnsureLineVisible(&m.viewport, m.state.Cursor)
		case "ctrl+u":
			halfPage := m.viewport.Height() / 2
			if halfPage < 1 {
				halfPage = 1
			}
			m.state.Cursor -= halfPage
			if m.state.Cursor < 0 {
				m.state.Cursor = 0
			}
			m.updateViewportContent()
			EnsureLineVisible(&m.viewport, m.state.Cursor)
		case "pgdown":
			fullPage := m.viewport.Height()
			if fullPage < 1 {
				fullPage = 1
			}
			m.state.Cursor += fullPage
			if m.state.Cursor >= len(m.state.VisibleItems) {
				m.state.Cursor = len(m.state.VisibleItems) - 1
			}
			if m.state.Cursor < 0 {
				m.state.Cursor = 0
			}
			m.updateViewportContent()
			EnsureLineVisible(&m.viewport, m.state.Cursor)
		case "pgup":
			fullPage := m.viewport.Height()
			if fullPage < 1 {
				fullPage = 1
			}
			m.state.Cursor -= fullPage
			if m.state.Cursor < 0 {
				m.state.Cursor = 0
			}
			m.updateViewportContent()
			EnsureLineVisible(&m.viewport, m.state.Cursor)

		case "s":
			m.sortBy = (m.sortBy + 1) % 3
			m.sortAgents(m.state.AllItems)
			m.applyContext()
			m.updateViewportContent()
		case "S":
			m.sortAsc = !m.sortAsc
			m.sortAgents(m.state.AllItems)
			m.applyContext()
			m.updateViewportContent()
		case "d":
			if len(m.state.VisibleItems) > 0 {
				return func() tea.Msg {
					return ShowAgentDetailMsg{Agent: m.state.VisibleItems[m.state.Cursor]}
				}
			}
		case "e", "E":
			if len(m.state.VisibleItems) > 0 {
				return func() tea.Msg {
					return EditAgentMsg{Agent: m.state.VisibleItems[m.state.Cursor]}
				}
			}
		}
	}

	return nil
}

func (m *AgentListModel) GetContext() Context {
	return m.state.Context
}

func (m *AgentListModel) SetContext(ctx Context) tea.Cmd {
	m.state.SetContext(ctx, m.agentHooks())
	m.updateViewportContent()
	return nil
}

func (m *AgentListModel) Reload() tea.Cmd {
	m.state.CaptureCursorForReload(m.agentHooks())
	m.state.Loading = true
	m.loadErr = nil
	return scanAgentsCmd
}

func (m *AgentListModel) View(width, height int) string {
	m.lastWidth = width

	if m.state.Loading {
		return renderCenteredText("Loading agents...", width, height)
	}
	if m.loadErr != nil {
		return renderCenteredText("Failed to load agents: "+m.loadErr.Error(), width, height)
	}
	if len(m.state.VisibleItems) == 0 {
		if m.state.Context.Type == ContextProject {
			return renderCenteredText("No agents found in project: "+m.state.Context.Value, width, height)
		}
		return renderCenteredText("No agents found", width, height)
	}

	return m.viewport.View()
}

// updateViewportContent updates the viewport with rendered agent table content
func (m *AgentListModel) updateViewportContent() {
	if m.state.Loading || m.loadErr != nil || len(m.state.VisibleItems) == 0 {
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

	contextLabel := ""
	if m.state.Context.Type == ContextAll {
		contextLabel = "All Agents"
	} else if m.state.Context.Type == ContextProject {
		contextLabel = m.state.Context.Value
	}
	content := renderAgentTable(m.state.VisibleItems, m.state.Cursor, width, height, m.ShowProjectColumn(), m.sortBy, m.sortAsc, contextLabel)
	m.viewport.SetContent(content)
}

func (m *AgentListModel) ApplyFilter(query string) {
	m.state.ApplyFilter(query, m.agentHooks())
	m.updateViewportContent()
}

func (m *AgentListModel) ShowProjectColumn() bool {
	return m.state.Context.Type == ContextAll
}

func (m *AgentListModel) captureCursorForReload() {
	m.state.CaptureCursorForReload(m.agentHooks())
}

func (m *AgentListModel) restoreCursorAfterReload() {
	m.state.RestoreCursorAfterReload(m.agentHooks())
}

func (m *AgentListModel) applyContext() {
	m.state.rebuild(m.agentHooks())
}

func (m *AgentListModel) applyFilter() {
	m.state.ApplyFilter(m.state.FilterQuery, m.agentHooks())
}

func agentMatchesQuery(agent claudefs.AgentResource, query string) bool {
	return strings.Contains(agentSearchText(agent), query)
}

func agentSearchText(agent claudefs.AgentResource) string {
	fields := []string{
		agent.Name,
		agent.Path,
		string(agent.Source),
		string(agent.Scope),
		agent.ProjectName,
		agent.PluginName,
		string(agent.Status),
		agent.Summary,
		agent.Description,
		agent.Model,
		agent.PermissionMode,
		agent.Memory,
		strings.Join(agent.Tools, " "),
		strings.Join(agent.ValidationReasons, " "),
		styles.AgentScopeText(agent.Scope),
	}

	fields = append(fields, agentSourceAliases(agent.Source)...)
	return strings.ToLower(strings.Join(fields, " "))
}

func agentSourceAliases(source claudefs.AgentSource) []string {
	switch source {
	case claudefs.AgentSourceProject:
		return []string{"project", "local"}
	case claudefs.AgentSourcePlugin:
		return []string{"plugin"}
	default:
		return []string{"user", "global"}
	}
}

func (m *AgentListModel) sortAgents(agents []claudefs.AgentResource) {
	if len(agents) == 0 {
		return
	}

	sort.SliceStable(agents, func(i, j int) bool {
		var less bool
		switch m.sortBy {
		case SortByAgentStatus:
			less = agents[i].Status < agents[j].Status
		case SortByAgentScope:
			less = agents[i].Scope < agents[j].Scope
		default:
			less = strings.ToLower(agents[i].Name) < strings.ToLower(agents[j].Name)
		}
		if m.sortAsc {
			return less
		}
		return !less
	})
}

func (m *AgentListModel) clampCursor() {
	m.state.ClampCursor()
}

func (m *AgentListModel) GetStats() (total, ready, invalid int) {
	for _, agent := range m.state.ContextItems {
		if agent.Status == claudefs.AgentStatusReady {
			ready++
		} else {
			invalid++
		}
	}
	return len(m.state.ContextItems), ready, invalid
}

func (m *AgentListModel) GetFilterStats() (filtered, total int) {
	return m.state.FilterStats()
}

func (m *AgentListModel) HasActiveFilter() bool {
	return m.state.HasActiveFilter()
}

func (m *AgentListModel) HasLoadError() bool {
	return m.loadErr != nil
}

func agentCursorKey(agent claudefs.AgentResource) string {
	return string(agent.Source) + "|" + agent.ProjectName + "|" + agent.PluginName + "|" + agent.Path
}

func agentAvailableInContext(agent claudefs.AgentResource, ctx Context) bool {
	if ctx.Type == ContextAll {
		return true
	}

	switch agent.Source {
	case claudefs.AgentSourceProject:
		return agent.ProjectName == ctx.Value
	case claudefs.AgentSourcePlugin:
		if agent.ProjectName == "" {
			return true
		}
		return agent.ProjectName == ctx.Value
	default:
		return true
	}
}

func (m *AgentListModel) GetSelectedAgent() (claudefs.AgentResource, bool) {
	if len(m.state.VisibleItems) == 0 || m.state.Cursor < 0 || m.state.Cursor >= len(m.state.VisibleItems) {
		return claudefs.AgentResource{}, false
	}
	return m.state.VisibleItems[m.state.Cursor], true
}

func (m *AgentListModel) agentHooks() DefaultResourceHooks[claudefs.AgentResource] {
	return DefaultResourceHooks[claudefs.AgentResource]{
		CursorKey:    agentCursorKey,
		InContext:    agentAvailableInContext,
		MatchesQuery: agentMatchesQuery,
	}
}
