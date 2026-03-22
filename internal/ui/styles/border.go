package styles

import "charm.land/lipgloss/v2"

var (
	PanelBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(ColorBorder)

	FocusBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(ColorBorderFocus)

	SeparatorStyle = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), true, false, false, false).
		BorderForeground(ColorBorder)

	CommandBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(ColorCommandBorder)

	SearchBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(ColorSearchBorder)
)
