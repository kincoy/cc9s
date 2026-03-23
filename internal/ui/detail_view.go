package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// DetailViewModel session detail panel Model
type DetailViewModel struct {
	session claudefs.Session       // session object
	stats   *claudefs.SessionStats // stats data (loaded async)
	loading bool                   // whether loading is in progress
	width   int                    // panel width
	height  int                    // panel height
}

// NewDetailViewModel creates a new detail panel Model
func NewDetailViewModel(session claudefs.Session) *DetailViewModel {
	return &DetailViewModel{
		session: session,
		loading: true,
	}
}

func (m *DetailViewModel) Init() tea.Cmd {
	return loadSessionStatsCmd(m.session)
}

// loadSessionStatsCmd asynchronously loads session stats
func loadSessionStatsCmd(session claudefs.Session) tea.Cmd {
	return func() tea.Msg {
		stats, err := claudefs.ParseSessionStats(session)
		return statsLoadedMsg{stats: stats, err: err}
	}
}

func (m *DetailViewModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case statsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			// Error handling: keep stats as nil
			m.stats = nil
		} else {
			m.stats = msg.stats
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			// Return to session list
			return func() tea.Msg {
				return CloseDetailMsg{}
			}
		}
	}

	return nil
}

func (m *DetailViewModel) View(width, height int) string {
	// Update dimensions
	m.width = width
	m.height = height

	// Generate panel with ViewBox, then center (for standalone rendering)
	box := m.ViewBox(width)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// ViewBox renders the detail panel box (without centering whitespace, for overlaying on background)
func (m *DetailViewModel) ViewBox(width int) string {
	// Calculate panel width: 60% of screen width, min 60 cols, max 100 cols
	panelWidth := int(float64(width) * 0.6)
	if panelWidth < 60 {
		panelWidth = 60
	}
	if panelWidth > 100 {
		panelWidth = 100
	}

	// If still loading
	if m.loading {
		content := styles.DetailTitleStyle.Render("Loading session details...")
		return lipgloss.NewStyle().
			Width(panelWidth).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorBorder).
			Render(content)
	}

	// If loading failed
	if m.stats == nil {
		content := styles.ErrorStyle.Render("Failed to load session details")
		return lipgloss.NewStyle().
			Width(panelWidth).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorBorder).
			Render(content)
	}

	// Render detail panel
	var sections []string

	// Title section
	title := fmt.Sprintf("Session Details: %s", m.session.ID[:8])
	sections = append(sections, styles.DetailTitleStyle.Render(title))
	sections = append(sections, "")

	// Metadata section
	sections = append(sections, m.renderMetadata())
	sections = append(sections, "")

	// Lifecycle section
	sections = append(sections, m.renderLifecycle())
	sections = append(sections, "")

	// Summary section
	sections = append(sections, m.renderSummary())
	sections = append(sections, "")

	// Dialog statistics section
	sections = append(sections, m.renderDialogStats())
	sections = append(sections, "")

	// Tool usage statistics section
	if len(m.stats.ToolUsage) > 0 {
		sections = append(sections, m.renderToolUsage())
		sections = append(sections, "")
	}

	// Token statistics section
	sections = append(sections, m.renderTokenStats())

	// Join all sections
	content := strings.Join(sections, "\n")

	// Apply panel style
	return lipgloss.NewStyle().
		Width(panelWidth).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Render(content)
}

// renderMetadata renders the metadata section
func (m *DetailViewModel) renderMetadata() string {
	var lines []string

	lines = append(lines, styles.DetailSectionStyle.Render("Metadata"))
	lines = append(lines, m.renderField("Session ID", m.session.ID))

	if m.stats.CustomTitle != "" {
		lines = append(lines, m.renderField("Title", m.stats.CustomTitle))
	}
	lines = append(lines, m.renderField("Model", m.stats.Model))
	lines = append(lines, m.renderField("Version", m.stats.Version))
	if m.stats.GitBranch != "" {
		lines = append(lines, m.renderField("Branch", m.stats.GitBranch))
	}
	lines = append(lines, m.renderField("Started", claudefs.FormatTime(m.stats.StartTime)))
	lines = append(lines, m.renderField("Last Active", claudefs.FormatTime(m.stats.LastActiveTime)))
	lines = append(lines, m.renderField("Duration", claudefs.FormatDuration(m.stats.Duration)))

	return strings.Join(lines, "\n")
}

// renderLifecycle renders the lifecycle state and explanation section.
func (m *DetailViewModel) renderLifecycle() string {
	var lines []string

	lifecycle := m.session.Lifecycle
	stateLabel := styles.LifecycleStatusStyle(lifecycle.State).Render(styles.LifecycleStatusText(lifecycle.State))

	lines = append(lines, styles.DetailSectionStyle.Render("Lifecycle"))
	lines = append(lines, m.renderField("State", stateLabel))

	for _, reason := range lifecycle.Evidence.Reasons {
		lines = append(lines, "  "+styles.DetailValueStyle.Render("- "+reason))
	}

	return strings.Join(lines, "\n")
}

// renderDialogStats renders the dialog statistics section
func (m *DetailViewModel) renderDialogStats() string {
	var lines []string

	lines = append(lines, styles.DetailSectionStyle.Render("Dialog Statistics"))
	lines = append(lines, m.renderField("Turns", fmt.Sprintf("%d", m.stats.TurnCount)))
	lines = append(lines, m.renderField("User Messages", fmt.Sprintf("%d", m.stats.UserMsgCount)))
	lines = append(lines, m.renderField("Tool Calls", fmt.Sprintf("%d", m.stats.ToolCallCount)))

	if m.stats.CompactCount > 0 {
		lines = append(lines, m.renderField("Compacts", fmt.Sprintf("%d", m.stats.CompactCount)))
	}
	if m.stats.ErrorCount > 0 {
		lines = append(lines, m.renderField("Errors", fmt.Sprintf("%d", m.stats.ErrorCount)))
	}

	return strings.Join(lines, "\n")
}

// renderToolUsage renders the tool usage statistics section
func (m *DetailViewModel) renderToolUsage() string {
	var lines []string

	lines = append(lines, styles.DetailSectionStyle.Render("Tool Usage (Top 5)"))

	// Sort by usage count (top 5)
	type toolCount struct {
		name  string
		count int
	}
	var tools []toolCount
	for name, count := range m.stats.ToolUsage {
		tools = append(tools, toolCount{name, count})
	}

	// Simple sort (descending)
	for i := 0; i < len(tools); i++ {
		for j := i + 1; j < len(tools); j++ {
			if tools[j].count > tools[i].count {
				tools[i], tools[j] = tools[j], tools[i]
			}
		}
	}

	// Take top 5
	limit := 5
	if len(tools) < limit {
		limit = len(tools)
	}

	for i := 0; i < limit; i++ {
		lines = append(lines, m.renderField(tools[i].name, fmt.Sprintf("%d", tools[i].count)))
	}

	return strings.Join(lines, "\n")
}

// renderTokenStats renders the token statistics section
func (m *DetailViewModel) renderTokenStats() string {
	var lines []string

	lines = append(lines, styles.DetailSectionStyle.Render("Token Statistics"))
	lines = append(lines, m.renderField("Input", claudefs.FormatNumber(m.stats.InputTokens)))
	lines = append(lines, m.renderField("Output", claudefs.FormatNumber(m.stats.OutputTokens)))
	if m.stats.CacheTokens > 0 {
		lines = append(lines, m.renderField("Cache", claudefs.FormatNumber(m.stats.CacheTokens)))
	}
	total := m.stats.InputTokens + m.stats.OutputTokens
	lines = append(lines, m.renderField("Total", claudefs.FormatNumber(total)))

	return strings.Join(lines, "\n")
}

// renderSummary renders the summary section (truncated to 150 chars, always shown)
func (m *DetailViewModel) renderSummary() string {
	var lines []string
	lines = append(lines, styles.DetailSectionStyle.Render("Summary"))
	summaryText := m.session.Summary
	if summaryText == "" {
		summaryText = "-"
	} else if utf8.RuneCountInString(summaryText) > 1000 {
		summaryText = string([]rune(summaryText)[:1000]) + "..."
	}
	lines = append(lines, styles.DetailValueStyle.Render(summaryText))
	return strings.Join(lines, "\n")
}

// renderField renders a single field (label + value)
func (m *DetailViewModel) renderField(label, value string) string {
	labelStyled := styles.DetailLabelStyle.Render(label + ":")
	valueStyled := styles.DetailValueStyle.Render(value)
	return fmt.Sprintf("  %s %s", labelStyled, valueStyled)
}
