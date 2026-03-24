package ui

import (
	"sort"
	"strings"

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
	context         Context
	agents          []claudefs.AgentResource
	allAgents       []claudefs.AgentResource
	contextAgents   []claudefs.AgentResource
	filterQuery     string
	cursor          int
	loading         bool
	loadErr         error
	sortBy          AgentSortField
	sortAsc         bool
	restoreAgentKey string
	restoreCursor   int
}

type agentsLoadedMsg struct {
	result claudefs.AgentScanResult
}

func NewAgentListModel() *AgentListModel {
	return &AgentListModel{
		context: Context{Type: ContextAll},
		loading: true,
		sortBy:  SortByAgentName,
		sortAsc: true,
	}
}

func (m *AgentListModel) Init() tea.Cmd {
	return scanAgentsCmd
}

func scanAgentsCmd() tea.Msg {
	return agentsLoadedMsg{result: claudefs.ScanAgents()}
}

func (m *AgentListModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case agentsLoadedMsg:
		m.loading = false
		m.loadErr = msg.result.Err
		m.allAgents = msg.result.Agents
		m.sortAgents(m.allAgents)
		m.applyContext()
		m.restoreCursorAfterReload()

	case tea.KeyPressMsg:
		if m.loadErr != nil {
			return nil
		}

		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.agents)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "G":
			if len(m.agents) > 0 {
				m.cursor = len(m.agents) - 1
			}
		case "g":
			m.cursor = 0
		case "s":
			m.sortBy = (m.sortBy + 1) % 3
			m.sortAgents(m.allAgents)
			m.applyContext()
		case "S":
			m.sortAsc = !m.sortAsc
			m.sortAgents(m.allAgents)
			m.applyContext()
		case "d":
			if len(m.agents) > 0 {
				return func() tea.Msg {
					return ShowAgentDetailMsg{Agent: m.agents[m.cursor]}
				}
			}
		case "e", "E":
			if len(m.agents) > 0 {
				return func() tea.Msg {
					return EditAgentMsg{Agent: m.agents[m.cursor]}
				}
			}
		}
	}

	return nil
}

func (m *AgentListModel) GetContext() Context {
	return m.context
}

func (m *AgentListModel) SetContext(ctx Context) tea.Cmd {
	m.context = ctx
	m.filterQuery = ""
	m.applyContext()
	return nil
}

func (m *AgentListModel) Reload() tea.Cmd {
	m.captureCursorForReload()
	m.loading = true
	m.loadErr = nil
	return scanAgentsCmd
}

func (m *AgentListModel) View(width, height int) string {
	if m.loading {
		return renderCenteredText("Loading agents...", width, height)
	}
	if m.loadErr != nil {
		return renderCenteredText("Failed to load agents: "+m.loadErr.Error(), width, height)
	}
	if len(m.agents) == 0 {
		if m.context.Type == ContextProject {
			return renderCenteredText("No agents found in project: "+m.context.Value, width, height)
		}
		return renderCenteredText("No agents found", width, height)
	}
	return renderAgentTable(m.agents, m.cursor, width, height, m.ShowProjectColumn(), m.sortBy, m.sortAsc)
}

func (m *AgentListModel) ApplyFilter(query string) {
	m.filterQuery = query
	m.applyFilter()
}

func (m *AgentListModel) ShowProjectColumn() bool {
	return m.context.Type == ContextAll
}

func (m *AgentListModel) captureCursorForReload() {
	m.restoreAgentKey = ""
	m.restoreCursor = m.cursor
	if m.cursor >= 0 && m.cursor < len(m.agents) {
		m.restoreAgentKey = agentCursorKey(m.agents[m.cursor])
	}
}

func (m *AgentListModel) restoreCursorAfterReload() {
	defer func() {
		m.restoreAgentKey = ""
		m.restoreCursor = 0
	}()

	if len(m.agents) == 0 {
		m.cursor = 0
		return
	}
	if m.restoreAgentKey != "" {
		for i, agent := range m.agents {
			if agentCursorKey(agent) == m.restoreAgentKey {
				m.cursor = i
				return
			}
		}
	}
	m.cursor = m.restoreCursor
	m.clampCursor()
}

func (m *AgentListModel) applyContext() {
	if m.context.Type == ContextAll {
		m.contextAgents = append([]claudefs.AgentResource(nil), m.allAgents...)
	} else {
		filtered := make([]claudefs.AgentResource, 0)
		for _, agent := range m.allAgents {
			if agentAvailableInContext(agent, m.context) {
				filtered = append(filtered, agent)
			}
		}
		m.contextAgents = filtered
	}
	m.applyFilter()
}

func (m *AgentListModel) applyFilter() {
	q := normalizeResourceSearchQuery(m.filterQuery)
	if q == "" {
		m.agents = append([]claudefs.AgentResource(nil), m.contextAgents...)
		m.clampCursor()
		return
	}

	filtered := make([]claudefs.AgentResource, 0)
	for _, agent := range m.contextAgents {
		if agentMatchesQuery(agent, q) {
			filtered = append(filtered, agent)
		}
	}
	m.agents = filtered
	m.clampCursor()
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
	if len(m.agents) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.agents) {
		m.cursor = len(m.agents) - 1
	}
}

func (m *AgentListModel) GetStats() (total, ready, invalid int) {
	for _, agent := range m.contextAgents {
		if agent.Status == claudefs.AgentStatusReady {
			ready++
		} else {
			invalid++
		}
	}
	return len(m.contextAgents), ready, invalid
}

func (m *AgentListModel) GetFilterStats() (filtered, total int) {
	return len(m.agents), len(m.contextAgents)
}

func (m *AgentListModel) HasActiveFilter() bool {
	return strings.TrimSpace(normalizeResourceSearchQuery(m.filterQuery)) != ""
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
	if len(m.agents) == 0 || m.cursor < 0 || m.cursor >= len(m.agents) {
		return claudefs.AgentResource{}, false
	}
	return m.agents[m.cursor], true
}
