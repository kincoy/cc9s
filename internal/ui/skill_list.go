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
	context         Context
	skills          []claudefs.SkillResource
	allSkills       []claudefs.SkillResource
	contextSkills   []claudefs.SkillResource
	filterQuery     string
	cursor          int
	loading         bool
	sortBy          SkillSortField
	sortAsc         bool
	readyCount      int
	invalidCount    int
	restoreSkillKey string
	restoreCursor   int
}

type skillsLoadedMsg struct {
	result claudefs.SkillScanResult
}

// NewSkillListModel creates a new skill list model.
func NewSkillListModel() *SkillListModel {
	return &SkillListModel{
		context: Context{Type: ContextAll},
		loading: true,
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
		m.loading = false
		m.allSkills = msg.result.Skills
		m.readyCount = msg.result.ReadyCount
		m.invalidCount = msg.result.InvalidCount
		m.sortSkills(m.allSkills)
		m.applyContext()
		m.restoreCursorAfterReload()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.skills)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "G":
			if len(m.skills) > 0 {
				m.cursor = len(m.skills) - 1
			}
		case "g":
			m.cursor = 0
		case "s":
			m.sortBy = (m.sortBy + 1) % 4
			m.sortSkills(m.allSkills)
			m.applyContext()
		case "S":
			m.sortAsc = !m.sortAsc
			m.sortSkills(m.allSkills)
			m.applyContext()
		case "d":
			if len(m.skills) > 0 {
				return func() tea.Msg {
					return ShowSkillDetailMsg{Skill: m.skills[m.cursor]}
				}
			}
		case "e", "E":
			if len(m.skills) > 0 {
				return func() tea.Msg {
					return EditSkillMsg{Skill: m.skills[m.cursor]}
				}
			}
		}
	}

	return nil
}

func (m *SkillListModel) GetContext() Context {
	return m.context
}

func (m *SkillListModel) SetContext(ctx Context) tea.Cmd {
	m.context = ctx
	m.filterQuery = ""
	m.applyContext()
	return nil
}

func (m *SkillListModel) Reload() tea.Cmd {
	m.captureCursorForReload()
	return scanSkillsCmd
}

func (m *SkillListModel) View(width, height int) string {
	if m.loading {
		return renderCenteredText("Loading skills...", width, height)
	}

	if len(m.skills) == 0 {
		if m.context.Type == ContextProject {
			return renderCenteredText(
				"No skills found in project: "+m.context.Value,
				width, height,
			)
		}
		return renderCenteredText("No skills found", width, height)
	}

	return renderSkillTable(m.skills, m.cursor, width, height, m.sortBy, m.sortAsc, m.ShowProjectColumn())
}

func (m *SkillListModel) ApplyFilter(query string) {
	m.filterQuery = query
	m.applyFilter()
}

func (m *SkillListModel) ShowProjectColumn() bool {
	return true
}

func (m *SkillListModel) captureCursorForReload() {
	m.restoreSkillKey = ""
	m.restoreCursor = m.cursor
	if m.cursor >= 0 && m.cursor < len(m.skills) {
		m.restoreSkillKey = skillCursorKey(m.skills[m.cursor])
	}
}

func (m *SkillListModel) restoreCursorAfterReload() {
	defer func() {
		m.restoreSkillKey = ""
		m.restoreCursor = 0
	}()

	if len(m.skills) == 0 {
		m.cursor = 0
		return
	}
	if m.restoreSkillKey != "" {
		for i, skill := range m.skills {
			if skillCursorKey(skill) == m.restoreSkillKey {
				m.cursor = i
				return
			}
		}
	}
	m.cursor = m.restoreCursor
	m.clampCursor()
}

func (m *SkillListModel) applyContext() {
	if m.context.Type == ContextAll {
		m.contextSkills = append([]claudefs.SkillResource(nil), m.allSkills...)
	} else {
		filtered := make([]claudefs.SkillResource, 0)
		for _, skill := range m.allSkills {
			if skillAvailableInContext(skill, m.context) {
				filtered = append(filtered, skill)
			}
		}
		m.contextSkills = filtered
	}
	m.applyFilter()
}

func (m *SkillListModel) applyFilter() {
	q := normalizeResourceSearchQuery(m.filterQuery)
	if q == "" {
		m.skills = append([]claudefs.SkillResource(nil), m.contextSkills...)
		m.clampCursor()
		return
	}

	var filtered []claudefs.SkillResource
	for _, skill := range m.contextSkills {
		if skillMatchesQuery(skill, q) {
			filtered = append(filtered, skill)
		}
	}
	m.skills = filtered
	m.clampCursor()
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
	if len(m.skills) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.skills) {
		m.cursor = len(m.skills) - 1
	}
}

func (m *SkillListModel) GetStats() (total, ready, invalid int) {
	for _, skill := range m.contextSkills {
		if skill.Status == claudefs.SkillStatusReady {
			ready++
		} else {
			invalid++
		}
	}
	return len(m.contextSkills), ready, invalid
}

func (m *SkillListModel) GetFilterStats() (filtered, total int) {
	return len(m.skills), len(m.contextSkills)
}

func (m *SkillListModel) HasActiveFilter() bool {
	return strings.TrimSpace(normalizeResourceSearchQuery(m.filterQuery)) != ""
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
	if len(m.skills) == 0 || m.cursor < 0 || m.cursor >= len(m.skills) {
		return claudefs.SkillResource{}, false
	}
	return m.skills[m.cursor], true
}
