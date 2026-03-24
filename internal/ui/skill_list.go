package ui

import (
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// SkillSortField skill sort field enum.
type SkillSortField int

const (
	SortBySkillName SkillSortField = iota
	SortBySkillType
	SortBySkillStatus
	SortBySkillScope
)

// SkillListModel skill list view model.
type SkillListModel struct {
	state   DefaultResourceListState[claudefs.SkillResource]
	sortBy  SkillSortField
	sortAsc bool
}

type skillsLoadedMsg struct {
	result claudefs.SkillScanResult
}

// NewSkillListModel creates a new skill list model.
func NewSkillListModel() *SkillListModel {
	return &SkillListModel{
		state:   NewDefaultResourceListState[claudefs.SkillResource](),
		sortBy:  SortBySkillName,
		sortAsc: true,
	}
}

func (m *SkillListModel) Init() tea.Cmd {
	return scanSkillsCmd
}

func scanSkillsCmd() tea.Msg {
	return skillsLoadedMsg{result: claudefs.ScanSkills()}
}

func (m *SkillListModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case skillsLoadedMsg:
		items := append([]claudefs.SkillResource(nil), msg.result.Skills...)
		m.sortSkills(items)
		m.state.SetItems(items, m.skillHooks())

	case tea.KeyPressMsg:
		switch msg.String() {
		case "j", "down":
			if m.state.Cursor < len(m.state.VisibleItems)-1 {
				m.state.Cursor++
			}
		case "k", "up":
			if m.state.Cursor > 0 {
				m.state.Cursor--
			}
		case "G":
			if len(m.state.VisibleItems) > 0 {
				m.state.Cursor = len(m.state.VisibleItems) - 1
			}
		case "g":
			m.state.Cursor = 0
		case "s":
			m.sortBy = (m.sortBy + 1) % 4
			m.sortSkills(m.state.AllItems)
			m.applyContext()
		case "S":
			m.sortAsc = !m.sortAsc
			m.sortSkills(m.state.AllItems)
			m.applyContext()
		case "d":
			if len(m.state.VisibleItems) > 0 {
				return func() tea.Msg {
					return ShowSkillDetailMsg{Skill: m.state.VisibleItems[m.state.Cursor]}
				}
			}
		case "e", "E":
			if len(m.state.VisibleItems) > 0 {
				return func() tea.Msg {
					return EditSkillMsg{Skill: m.state.VisibleItems[m.state.Cursor]}
				}
			}
		}
	}

	return nil
}

func (m *SkillListModel) GetContext() Context {
	return m.state.Context
}

func (m *SkillListModel) SetContext(ctx Context) tea.Cmd {
	m.state.SetContext(ctx, m.skillHooks())
	return nil
}

func (m *SkillListModel) Reload() tea.Cmd {
	m.state.CaptureCursorForReload(m.skillHooks())
	m.state.Loading = true
	return scanSkillsCmd
}

func (m *SkillListModel) View(width, height int) string {
	if m.state.Loading {
		return renderCenteredText("Loading skills...", width, height)
	}

	if len(m.state.VisibleItems) == 0 {
		if m.state.Context.Type == ContextProject {
			return renderCenteredText(
				"No skills found in project: "+m.state.Context.Value,
				width, height,
			)
		}
		return renderCenteredText("No skills found", width, height)
	}

	return renderSkillTable(m.state.VisibleItems, m.state.Cursor, width, height, m.sortBy, m.sortAsc, m.ShowProjectColumn())
}

func (m *SkillListModel) ApplyFilter(query string) {
	m.state.ApplyFilter(query, m.skillHooks())
}

func (m *SkillListModel) ShowProjectColumn() bool {
	return true
}

func (m *SkillListModel) captureCursorForReload() {
	m.state.CaptureCursorForReload(m.skillHooks())
}

func (m *SkillListModel) restoreCursorAfterReload() {
	m.state.RestoreCursorAfterReload(m.skillHooks())
}

func (m *SkillListModel) applyContext() {
	m.state.rebuild(m.skillHooks())
}

func (m *SkillListModel) applyFilter() {
	m.state.ApplyFilter(m.state.FilterQuery, m.skillHooks())
}

func skillMatchesQuery(skill claudefs.SkillResource, query string) bool {
	return strings.Contains(skillSearchText(skill), query)
}

func skillSearchText(skill claudefs.SkillResource) string {
	fields := []string{
		skill.Name,
		skill.Path,
		string(skill.Kind),
		string(skill.Source),
		string(skill.Scope),
		skill.ProjectName,
		skill.PluginName,
		string(skill.Status),
		skill.Summary,
		strings.Join(skill.ValidationReasons, " "),
		styles.SkillScopeText(skill.Scope),
	}

	fields = append(fields, skillSourceAliases(skill.Source)...)
	fields = append(fields, skillKindAliases(skill.Kind)...)

	return strings.ToLower(strings.Join(fields, " "))
}

func skillSourceAliases(source claudefs.SkillSource) []string {
	switch source {
	case claudefs.SkillSourceProject:
		return []string{"project", "local"}
	case claudefs.SkillSourcePlugin:
		return []string{"plugin"}
	default:
		return []string{"user", "global"}
	}
}

func skillKindAliases(kind claudefs.SkillKind) []string {
	switch kind {
	case claudefs.SkillKindCommand:
		return []string{"command", "cmd"}
	default:
		return []string{"skill"}
	}
}

func (m *SkillListModel) sortSkills(skills []claudefs.SkillResource) {
	if len(skills) == 0 {
		return
	}

	sort.SliceStable(skills, func(i, j int) bool {
		var less bool
		switch m.sortBy {
		case SortBySkillScope:
			less = skills[i].Scope < skills[j].Scope
		case SortBySkillType:
			less = strings.ToLower(string(skills[i].Kind)) < strings.ToLower(string(skills[j].Kind))
		case SortBySkillStatus:
			less = skills[i].Status < skills[j].Status
		default:
			less = strings.ToLower(skills[i].Name) < strings.ToLower(skills[j].Name)
		}
		if m.sortAsc {
			return less
		}
		return !less
	})
}

func (m *SkillListModel) clampCursor() {
	m.state.ClampCursor()
}

func (m *SkillListModel) GetStats() (total, ready, invalid int) {
	for _, skill := range m.state.ContextItems {
		if skill.Status == claudefs.SkillStatusReady {
			ready++
		} else {
			invalid++
		}
	}
	return len(m.state.ContextItems), ready, invalid
}

func (m *SkillListModel) GetFilterStats() (filtered, total int) {
	return m.state.FilterStats()
}

func (m *SkillListModel) HasActiveFilter() bool {
	return m.state.HasActiveFilter()
}

func skillCursorKey(skill claudefs.SkillResource) string {
	return string(skill.Source) + "|" + skill.ProjectName + "|" + skill.Path
}

func skillAvailableInContext(skill claudefs.SkillResource, ctx Context) bool {
	if ctx.Type == ContextAll {
		return true
	}

	switch skill.Source {
	case claudefs.SkillSourceProject:
		return skill.ProjectName == ctx.Value
	case claudefs.SkillSourcePlugin:
		if skill.ProjectName == "" {
			return true
		}
		return skill.ProjectName == ctx.Value
	default:
		return true
	}
}

func (m *SkillListModel) GetSelectedSkill() (claudefs.SkillResource, bool) {
	if len(m.state.VisibleItems) == 0 || m.state.Cursor < 0 || m.state.Cursor >= len(m.state.VisibleItems) {
		return claudefs.SkillResource{}, false
	}
	return m.state.VisibleItems[m.state.Cursor], true
}

func (m *SkillListModel) skillHooks() DefaultResourceHooks[claudefs.SkillResource] {
	return DefaultResourceHooks[claudefs.SkillResource]{
		CursorKey:    skillCursorKey,
		InContext:    skillAvailableInContext,
		MatchesQuery: skillMatchesQuery,
	}
}
