package styles

import "charm.land/lipgloss/v2"

var (
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

	MultiSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1A3A2A")).
				Foreground(lipgloss.Color("#FFFFFF"))

	TableCellStyle = lipgloss.NewStyle().
			Foreground(ColorNormal)

	ActiveStatusStyle = ActiveStyle

	IdleStatusStyle = IdleStyle

	CompletedStatusStyle = CompletedStyle

	StaleStatusStyle = StaleStyle

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
)
