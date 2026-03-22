package styles

import "charm.land/lipgloss/v2"

var (
	ActiveStyle = lipgloss.NewStyle().
		Foreground(ColorActive).
		Bold(true)

	WarningStyle = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true)
)
