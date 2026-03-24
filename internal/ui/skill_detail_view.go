package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// SkillDetailViewModel skill detail panel model.
type SkillDetailViewModel struct {
	skill  claudefs.SkillResource
	width  int
	height int
}

// NewSkillDetailViewModel creates a new skill detail panel.
func NewSkillDetailViewModel(skill claudefs.SkillResource) *SkillDetailViewModel {
	return &SkillDetailViewModel{skill: skill}
}

func (m *SkillDetailViewModel) Init() tea.Cmd {
	return nil
}

func (m *SkillDetailViewModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		if msg.String() == "esc" {
			return func() tea.Msg { return CloseSkillDetailMsg{} }
		}
	}

	return nil
}

func (m *SkillDetailViewModel) View(width, height int) string {
	m.width = width
	m.height = height
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, m.ViewBox(width))
}

func (m *SkillDetailViewModel) ViewBox(width int) string {
	panelWidth := int(float64(width) * 0.6)
	if panelWidth < 60 {
		panelWidth = 60
	}
	if panelWidth > 100 {
		panelWidth = 100
	}

	sections := []string{
		styles.DetailTitleStyle.Render(fmt.Sprintf("Skill Details: %s", m.skill.Name)),
		"",
		m.renderMetadata(),
		"",
		m.renderAvailability(),
	}

	if m.skill.Summary != "" {
		sections = append(sections, "", m.renderSummary())
	}

	if len(m.skill.AssociatedFiles) > 0 {
		sections = append(sections, "", m.renderAssociatedFiles(panelWidth))
	}

	content := strings.Join(sections, "\n")
	return lipgloss.NewStyle().
		Width(panelWidth).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Render(content)
}

func (m *SkillDetailViewModel) renderMetadata() string {
	lines := []string{
		styles.DetailSectionStyle.Render("Metadata"),
		m.renderField("Name", m.skill.Name),
		m.renderField("Type", string(m.skill.Kind)),
		m.renderField("Source", string(m.skill.Source)),
		m.renderField("Scope", string(m.skill.Scope)),
		m.renderField("Shape", string(m.skill.Shape)),
		m.renderField("Path", m.skill.Path),
		m.renderField("Entry File", entryDisplayPath(m.skill)),
	}
	if m.skill.ProjectName != "" {
		lines = append(lines, m.renderField("Project", m.skill.ProjectName))
	}
	if m.skill.PluginName != "" {
		lines = append(lines, m.renderField("Plugin", m.skill.PluginName))
	}
	if m.skill.PluginInstallMode != "" {
		lines = append(lines, m.renderField("Plugin Scope", m.skill.PluginInstallMode))
	}
	return strings.Join(lines, "\n")
}

func (m *SkillDetailViewModel) renderAvailability() string {
	lines := []string{
		styles.DetailSectionStyle.Render("Availability"),
		m.renderField("Status", styles.SkillStatusStyle(m.skill.Status).Render(styles.SkillStatusText(m.skill.Status))),
	}

	for _, reason := range m.skill.ValidationReasons {
		lines = append(lines, "  "+styles.DetailValueStyle.Render("- "+reason))
	}
	return strings.Join(lines, "\n")
}

func (m *SkillDetailViewModel) renderSummary() string {
	lines := []string{
		styles.DetailSectionStyle.Render("Description"),
		styles.DetailValueStyle.Render("Path: " + m.skill.Path),
		"",
		styles.DetailValueStyle.Render(m.skill.Summary),
	}
	return strings.Join(lines, "\n")
}

func (m *SkillDetailViewModel) renderAssociatedFiles(panelWidth int) string {
	lines := []string{styles.DetailSectionStyle.Render("Associated Files")}
	limit := 6
	for i, file := range m.skill.AssociatedFiles {
		if i >= limit {
			lines = append(lines, styles.DetailValueStyle.Render(fmt.Sprintf("  ... and %d more", len(m.skill.AssociatedFiles)-limit)))
			break
		}
		lines = append(lines, styles.DetailValueStyle.Render("  - "+clampDisplayText(file, panelWidth-8)))
	}
	return strings.Join(lines, "\n")
}

func (m *SkillDetailViewModel) renderField(label, value string) string {
	labelStyled := styles.DetailLabelStyle.Render(label + ":")
	valueStyled := styles.DetailValueStyle.Render(value)
	return labelStyled + " " + valueStyled
}

func entryDisplayPath(skill claudefs.SkillResource) string {
	if skill.Shape == claudefs.SkillShapeSingleFile {
		return filepath.Base(skill.EntryFile)
	}
	rel, err := filepath.Rel(skill.Path, skill.EntryFile)
	if err != nil {
		return skill.EntryFile
	}
	return rel
}
