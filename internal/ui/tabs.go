package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// TabsModel manages the tab navigation bar
type TabsModel struct {
	current ResourceType
}

var tabResources = []ResourceType{
	ResourceProjects,
	ResourceSessions,
	ResourceSkills,
	ResourceAgents,
}

func NewTabsModel() *TabsModel {
	return &TabsModel{
		current: ResourceProjects,
	}
}

func (t *TabsModel) SetCurrent(r ResourceType) {
	t.current = r
}

// TabsHeight is the vertical space consumed by the tab bar.
const TabsHeight = 2

func (t *TabsModel) Render(width int) string {
	var parts []string
	for _, r := range tabResources {
		name := tabDisplayName(r)
		if r == t.current {
			style := lipgloss.NewStyle().
				Bold(true).
				Underline(true).
				Foreground(styles.ColorHighlight)
			parts = append(parts, style.Render(name))
		} else {
			parts = append(parts, lipgloss.NewStyle().
				Foreground(styles.ColorDim).
				Render(name))
		}
	}

	line := strings.Join(parts, "  ")
	tabLine := lipgloss.NewStyle().
		Width(width).
		Foreground(styles.ColorDim).
		Render(line)

	// Horizontal separator below tabs (matches TabsHeight = 2)
	separator := lipgloss.NewStyle().
		Width(width).
		Foreground(styles.ColorBorder).
		Render(strings.Repeat("─", width))

	return tabLine + "\n" + separator
}

func tabDisplayName(r ResourceType) string {
	switch r {
	case ResourceProjects:
		return "Projects"
	case ResourceSessions:
		return "Sessions"
	case ResourceSkills:
		return "Skills"
	case ResourceAgents:
		return "Agents"
	default:
		return "Unknown"
	}
}
