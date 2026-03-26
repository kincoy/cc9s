package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

var (
	ColorActive    = lipgloss.Color("#04B575")
	ColorIdle      = lipgloss.Color("#FBBF24")
	ColorCompleted = lipgloss.Color("#60A5FA")
	ColorStale     = lipgloss.Color("#EF4444")
	ColorNormal    = lipgloss.Color("#D4D4D4")
	ColorWarning   = lipgloss.Color("#FBBF24")
	ColorError     = lipgloss.Color("#EF4444")
	ColorDim       = lipgloss.Color("#888888")
)

var (
	ColorHighlight = lipgloss.Color("#7DCFFF")
	ColorTitle     = lipgloss.Color("#FF6600")
	ColorPurple    = lipgloss.Color("#7D56F4")
)

var (
	ColorBorder      = lipgloss.Color("#3A3A5A")
	ColorBorderFocus = lipgloss.Color("#7DCFFF")
)

var (
	ColorCommandBorder = lipgloss.Color("#7DCFFF")
	ColorSearchBorder  = lipgloss.Color("#2DD4BF")
)

var (
	ColorFlashSuccess  = lipgloss.Color("#10B981")
	ColorTableHeaderBg = lipgloss.Color("#1a1a2e")
)

// ColorBackground holds the optional background color for the full TUI.
// When nil (zero value), no background is applied.
var ColorBackground color.Color

// ForceBackground controls whether ColorBackground is applied to the root view.
var ForceBackground bool

// rebuildStyles reconstructs every package-level lipgloss.Style variable
// from the current color variables. This is necessary because lipgloss styles
// capture color values at construction time; changing a color variable alone
// does not retroactively update existing styles.
func rebuildStyles() {
	// --- text.go ---
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

	// --- styles.go ---
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorTitle)

	BreadcrumbStyle = lipgloss.NewStyle().
		Foreground(ColorDim)

	FooterStyle = lipgloss.NewStyle().
		Foreground(ColorDim)

	FooterKeyStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorHighlight)

	BodyStyle = lipgloss.NewStyle()

	TableHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorHighlight)

	TableCellStyle = lipgloss.NewStyle().
		Foreground(ColorNormal)

	MultiSelectedStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#1A3A2A")).
		Foreground(lipgloss.Color("#FFFFFF"))

	ActiveStatusStyle = lipgloss.NewStyle().
		Foreground(ColorActive).
		Bold(true)

	IdleStatusStyle = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true)

	CompletedStatusStyle = lipgloss.NewStyle().
		Foreground(ColorCompleted).
		Bold(true)

	StaleStatusStyle = lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true)

	ReadySkillStatusStyle = lipgloss.NewStyle().
		Foreground(ColorActive).
		Bold(true)

	InvalidSkillStatusStyle = lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true)

	ReadyAgentStatusStyle = lipgloss.NewStyle().
		Foreground(ColorActive).
		Bold(true)

	InvalidAgentStatusStyle = lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true)

	ProjectSkillSourceStyle = lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Bold(true)

	UserSkillSourceStyle = lipgloss.NewStyle().
		Foreground(ColorCompleted).
		Bold(true)

	PluginSkillSourceStyle = lipgloss.NewStyle().
		Foreground(ColorPurple).
		Bold(true)

	ProjectAgentSourceStyle = lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Bold(true)

	UserAgentSourceStyle = lipgloss.NewStyle().
		Foreground(ColorCompleted).
		Bold(true)

	PluginAgentSourceStyle = lipgloss.NewStyle().
		Foreground(ColorPurple).
		Bold(true)

	SkillKindBadgeStyle = lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Bold(true)

	CommandKindBadgeStyle = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true)

	DialogTitleStyle = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true)

	DialogMessageStyle = lipgloss.NewStyle().
		Foreground(ColorNormal)

	DialogButtonStyle = lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Bold(true)

	DetailTitleStyle = lipgloss.NewStyle().
		Foreground(ColorTitle).
		Bold(true)

	DetailSectionStyle = lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Bold(true).
		Underline(true)

	DetailLabelStyle = lipgloss.NewStyle().
		Foreground(ColorDim).
		Bold(true)

	DetailValueStyle = lipgloss.NewStyle().
		Foreground(ColorNormal)

	LogTitleStyle = lipgloss.NewStyle().
		Foreground(ColorTitle).
		Bold(true)

	LogTurnHeaderStyle = lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Bold(true)

	LogUserStyle = lipgloss.NewStyle().
		Foreground(ColorActive).
		Bold(true)

	LogAssistantStyle = lipgloss.NewStyle().
		Foreground(ColorNormal).
		Bold(true)

	LogToolStyle = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true)

	LogHintStyle = lipgloss.NewStyle().
		Foreground(ColorDim).
		Italic(true)

	// --- status.go ---
	ActiveStyle = lipgloss.NewStyle().
		Foreground(ColorActive).
		Bold(true)

	IdleStyle = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true)

	CompletedStyle = lipgloss.NewStyle().
		Foreground(ColorCompleted).
		Bold(true)

	StaleStyle = lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true)

	WarningStyle = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true)

	// --- border.go ---
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
}
