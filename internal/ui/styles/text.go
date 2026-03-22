package styles

import "charm.land/lipgloss/v2"

var (
	TitleStyle = lipgloss.NewStyle().
		Foreground(ColorTitle).
		Bold(true)

	HighlightStyle = lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Bold(true)

	NormalStyle = lipgloss.NewStyle().
		Foreground(ColorNormal)

	DimStyle = lipgloss.NewStyle().
		Foreground(ColorDim)

	SelectedRowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(ColorHighlight).
		Bold(true)
)
