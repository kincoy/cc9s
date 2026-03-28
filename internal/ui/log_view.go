package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// LogViewModel session log view Model
type LogViewModel struct {
	session    claudefs.Session      // session object
	logEntries []claudefs.LogEntry   // log entries (organized by turn)
	loading    bool                  // whether loading is in progress
	offset     int                   // current offset
	hasMore    bool                  // whether more log entries exist
	viewport   viewport.Model       // viewport for scrolling
	width      int                   // view width
	height     int                   // view height
}

// NewLogViewModel creates a new log view Model
func NewLogViewModel(session claudefs.Session) *LogViewModel {
	return &LogViewModel{
		session: session,
		loading: true,
		offset:  0,
	}
}

func (m *LogViewModel) Init() tea.Cmd {
	m.viewport = NewViewportWithSize(80, 20)
	return tea.Batch(
		m.viewport.Init(),
		loadSessionLogCmd(m.session.ID, 0, 100),
	)
}

// loadSessionLogCmd asynchronously loads session log
func loadSessionLogCmd(sessionID string, offset, limit int) tea.Cmd {
	return func() tea.Msg {
		entries, total, err := claudefs.ParseSessionLog(sessionID, offset, limit)
		return logLoadedMsg{entries: entries, total: total, err: err}
	}
}

func (m *LogViewModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case logLoadedMsg:
		m.loading = false
		if msg.err != nil {
			// Error handling: keep empty list
			m.logEntries = nil
		} else {
			m.logEntries = msg.entries
			m.hasMore = len(msg.entries) < msg.total
		}
		m.updateViewportContent()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Account for padding (1 top + 1 bottom = 2 vertical, 2 left + 2 right = 4 horizontal)
		vpWidth := msg.Width - 4
		vpHeight := msg.Height - 2
		if vpWidth < 1 {
			vpWidth = 1
		}
		if vpHeight < 1 {
			vpHeight = 1
		}
		m.viewport.SetWidth(vpWidth)
		m.viewport.SetHeight(vpHeight)
		m.updateViewportContent()

	case tea.KeyPressMsg:
		switch msg.String() {
		// Scroll shortcuts
		case "j", "down":
			m.viewport.ScrollDown(1)
		case "k", "up":
			m.viewport.ScrollUp(1)
		case "G":
			m.viewport.GotoBottom()
		case "g":
			m.viewport.GotoTop()
		case "pgup":
			m.viewport.PageUp()
		case "pgdown":
			m.viewport.PageDown()
		case "ctrl+d":
			m.viewport.HalfPageDown()
		case "ctrl+u":
			m.viewport.HalfPageUp()

		// Go back
		case "esc":
			return func() tea.Msg {
				return CloseLogMsg{}
			}
		}
	}

	return nil
}

func (m *LogViewModel) View(width, height int) string {
	// Update dimensions
	m.width = width
	m.height = height

	// If still loading
	if m.loading {
		content := styles.LogTitleStyle.Render("Loading session log...")
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
	}

	// If loading failed
	if m.logEntries == nil {
		content := styles.ErrorStyle.Render("Failed to load session log")
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
	}

	// If no log entries
	if len(m.logEntries) == 0 {
		content := styles.LogTitleStyle.Render("No log entries found")
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
	}

	// Apply container style (fullscreen fill) around viewport
	container := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 2).
		Render(m.viewport.View())

	return container
}

// updateViewportContent renders all log entries and sets viewport content.
func (m *LogViewModel) updateViewportContent() {
	if m.loading {
		m.viewport.SetContent(styles.LogTitleStyle.Render("Loading session log..."))
		return
	}
	if m.logEntries == nil {
		m.viewport.SetContent(styles.ErrorStyle.Render("Failed to load session log"))
		return
	}
	if len(m.logEntries) == 0 {
		m.viewport.SetContent(styles.LogTitleStyle.Render("No log entries found"))
		return
	}

	var lines []string

	// Title
	title := fmt.Sprintf("Session Log: %s (%d turns)", m.session.ID[:8], len(m.logEntries))
	lines = append(lines, styles.LogTitleStyle.Render(title))
	lines = append(lines, "")

	// Render ALL log entries
	for i := range m.logEntries {
		entry := m.logEntries[i]
		lines = append(lines, m.renderLogEntry(entry))
		lines = append(lines, "")
	}

	// Hint message
	if m.hasMore {
		lines = append(lines, styles.LogHintStyle.Render("(More entries available)"))
	}

	content := strings.Join(lines, "\n")
	m.viewport.SetContent(content)
}

// renderLogEntry renders a single log entry (one turn)
func (m *LogViewModel) renderLogEntry(entry claudefs.LogEntry) string {
	var lines []string

	// Turn header
	header := fmt.Sprintf("Turn #%d  %s  (%dms)",
		entry.TurnNumber,
		claudefs.FormatTime(entry.Timestamp),
		entry.Duration,
	)
	lines = append(lines, styles.LogTurnHeaderStyle.Render(header))

	// User message
	if entry.UserMsg != "" {
		userLabel := styles.LogUserStyle.Render("User:")
		userMsg := m.truncateText(entry.UserMsg, 200)
		lines = append(lines, fmt.Sprintf("  %s %s", userLabel, userMsg))
	}

	// Assistant message
	if entry.AssistantMsg != "" {
		assistantLabel := styles.LogAssistantStyle.Render("Assistant:")
		assistantMsg := m.truncateText(entry.AssistantMsg, 200)
		lines = append(lines, fmt.Sprintf("  %s %s", assistantLabel, assistantMsg))
	}

	// Tool calls
	if len(entry.ToolCalls) > 0 {
		toolLabel := styles.LogToolStyle.Render("Tools:")
		lines = append(lines, fmt.Sprintf("  %s", toolLabel))
		for _, tool := range entry.ToolCalls {
			toolLine := fmt.Sprintf("    - %s: %s", tool.Name, m.truncateText(tool.Output, 80))
			if tool.IsError {
				toolLine = styles.ErrorStyle.Render(toolLine)
			}
			lines = append(lines, toolLine)
		}
	}

	return strings.Join(lines, "\n")
}

// truncateText truncates text and appends ellipsis (UTF-8 safe)
func (m *LogViewModel) truncateText(text string, maxRunes int) string {
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	return string(runes[:maxRunes]) + "..."
}
