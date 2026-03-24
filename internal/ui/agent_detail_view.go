package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// AgentDetailViewModel agent detail panel model.
type AgentDetailViewModel struct {
	agent  claudefs.AgentResource
	width  int
	height int
}

func NewAgentDetailViewModel(agent claudefs.AgentResource) *AgentDetailViewModel {
	return &AgentDetailViewModel{agent: agent}
}

func (m *AgentDetailViewModel) Init() tea.Cmd {
	return nil
}

func (m *AgentDetailViewModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		if msg.String() == "esc" {
			return func() tea.Msg { return CloseAgentDetailMsg{} }
		}
	}

	return nil
}

func (m *AgentDetailViewModel) View(width, height int) string {
	m.width = width
	m.height = height
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, m.ViewBox(width))
}

func (m *AgentDetailViewModel) ViewBox(width int) string {
	panelWidth := int(float64(width) * 0.6)
	if panelWidth < 60 {
		panelWidth = 60
	}
	if panelWidth > 100 {
		panelWidth = 100
	}

	sections := []string{
		styles.DetailTitleStyle.Render(fmt.Sprintf("Agent Details: %s", m.agent.Name)),
		"",
		m.renderMetadata(),
		"",
		m.renderAvailability(),
	}

	if description := strings.TrimSpace(m.agent.Description); description != "" || strings.TrimSpace(m.agent.Summary) != "" {
		sections = append(sections, "", m.renderDescription())
	}

	content := strings.Join(sections, "\n")
	return lipgloss.NewStyle().
		Width(panelWidth).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Render(content)
}

func (m *AgentDetailViewModel) renderMetadata() string {
	lines := []string{
		styles.DetailSectionStyle.Render("Metadata"),
		m.renderField("Name", m.agent.Name),
		m.renderField("Source", string(m.agent.Source)),
		m.renderField("Scope", string(m.agent.Scope)),
		m.renderField("Path", m.agent.Path),
	}
	if m.agent.ProjectName != "" {
		lines = append(lines, m.renderField("Project", m.agent.ProjectName))
	}
	if m.agent.PluginName != "" {
		lines = append(lines, m.renderField("Plugin", m.agent.PluginName))
	}
	if m.agent.Model != "" {
		lines = append(lines, m.renderField("Model", m.agent.Model))
	}
	if len(m.agent.Tools) > 0 {
		lines = append(lines, m.renderField("Tools", strings.Join(m.agent.Tools, ", ")))
	}
	if m.agent.PermissionMode != "" {
		lines = append(lines, m.renderField("Permission", m.agent.PermissionMode))
	}
	if m.agent.Memory != "" {
		lines = append(lines, m.renderField("Memory", m.agent.Memory))
	}
	return strings.Join(lines, "\n")
}

func (m *AgentDetailViewModel) renderAvailability() string {
	lines := []string{
		styles.DetailSectionStyle.Render("Availability"),
		m.renderField("Status", styles.AgentStatusStyle(m.agent.Status).Render(styles.AgentStatusText(m.agent.Status))),
	}
	for _, reason := range m.agent.ValidationReasons {
		lines = append(lines, "  "+styles.DetailValueStyle.Render("- "+reason))
	}
	return strings.Join(lines, "\n")
}

func (m *AgentDetailViewModel) renderDescription() string {
	description := strings.TrimSpace(m.agent.Description)
	if description == "" {
		description = strings.TrimSpace(m.agent.Summary)
	}

	lines := []string{
		styles.DetailSectionStyle.Render("Description"),
		styles.DetailValueStyle.Render(description),
	}
	return strings.Join(lines, "\n")
}

func (m *AgentDetailViewModel) renderField(label, value string) string {
	labelStyled := styles.DetailLabelStyle.Render(label + ":")
	valueStyled := styles.DetailValueStyle.Render(value)
	return labelStyled + " " + valueStyled
}
