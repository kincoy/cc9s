package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// ProjectDetailViewModel project detail panel model.
type ProjectDetailViewModel struct {
	project claudefs.Project
	width   int
	height  int
}

func NewProjectDetailViewModel(project claudefs.Project) *ProjectDetailViewModel {
	return &ProjectDetailViewModel{project: project}
}

func (m *ProjectDetailViewModel) Init() tea.Cmd {
	return nil
}

func (m *ProjectDetailViewModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		if msg.String() == "esc" {
			return func() tea.Msg { return CloseProjectDetailMsg{} }
		}
	}

	return nil
}

func (m *ProjectDetailViewModel) View(width, height int) string {
	m.width = width
	m.height = height
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, m.ViewBox(width))
}

func (m *ProjectDetailViewModel) ViewBox(width int) string {
	panelWidth := int(float64(width) * 0.65)
	if panelWidth < 64 {
		panelWidth = 64
	}
	if panelWidth > 104 {
		panelWidth = 104
	}

	sections := []string{
		styles.DetailTitleStyle.Render(fmt.Sprintf("Project Details: %s", m.project.Name)),
		"",
		m.renderMetadata(),
		"",
		m.renderResources(),
		"",
		m.renderRoots(),
	}

	content := strings.Join(sections, "\n")
	return lipgloss.NewStyle().
		Width(panelWidth).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Render(content)
}

func (m *ProjectDetailViewModel) renderMetadata() string {
	lines := []string{
		styles.DetailSectionStyle.Render("Metadata"),
		m.renderField("Name", m.project.Name),
		m.renderField("Path", m.project.Path),
		m.renderField("Encoded Path", m.project.EncodedPath),
		m.renderField("Last Active", claudefs.FormatTime(m.project.LastActiveAt)),
		m.renderField("Total Size", claudefs.FormatSize(m.project.TotalSize)),
	}
	return strings.Join(lines, "\n")
}

func (m *ProjectDetailViewModel) renderResources() string {
	lines := []string{
		styles.DetailSectionStyle.Render("Resources"),
		m.renderField("Sessions", fmt.Sprintf("%d total / %d active", m.project.SessionCount, m.project.ActiveSessionCount)),
		m.renderField("Local Skills", fmt.Sprintf("%d", m.project.SkillCount)),
		m.renderField("Local Commands", fmt.Sprintf("%d", m.project.CommandCount)),
		m.renderField("Local Agents", fmt.Sprintf("%d", m.project.AgentCount)),
		m.renderField("Context Note", "Context views include global user/plugin resources"),
	}
	return strings.Join(lines, "\n")
}

func (m *ProjectDetailViewModel) renderRoots() string {
	lines := []string{
		styles.DetailSectionStyle.Render("Claude Roots"),
		m.renderField(".claude/skills", projectRootStatus(m.project.HasSkillsRoot, m.project.SkillCount)),
		m.renderField(".claude/commands", projectRootStatus(m.project.HasCommandsRoot, m.project.CommandCount)),
		m.renderField(".claude/agents", projectRootStatus(m.project.HasAgentsRoot, m.project.AgentCount)),
	}
	return strings.Join(lines, "\n")
}

func projectRootStatus(exists bool, count int) string {
	if !exists {
		return "Missing"
	}
	return fmt.Sprintf("Present (%d)", count)
}

func (m *ProjectDetailViewModel) renderField(label, value string) string {
	labelStyled := styles.DetailLabelStyle.Render(label + ":")
	valueStyled := styles.DetailValueStyle.Render(value)
	return labelStyled + " " + valueStyled
}
