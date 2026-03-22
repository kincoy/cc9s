package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// LogViewModel session log view Model
type LogViewModel struct {
	session    claudefs.Session      // session object
	logEntries []claudefs.LogEntry   // log entries (organized by turn)
	loading    bool              // whether loading is in progress
	offset     int               // current offset
	hasMore    bool              // whether more log entries exist
	cursor     int               // current scroll position (first visible turn)
	width      int               // view width
	height     int               // view height
}

// NewLogViewModel creates a new log view Model
func NewLogViewModel(session claudefs.Session) *LogViewModel {
	return &LogViewModel{
		session: session,
		loading: true,
		offset:  0,
		cursor:  0,
	}
}

func (m *LogViewModel) Init() tea.Cmd {
	return loadSessionLogCmd(m.session.ID, 0, 100)
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

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch msg.String() {
		// Scroll shortcuts
		case "j", "down":
			if m.cursor < len(m.logEntries)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "G":
			if len(m.logEntries) > 0 {
				m.cursor = len(m.logEntries) - 1
			}
		case "g":
			m.cursor = 0

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

	// Render log view (fullscreen)
	var lines []string

	// Title
	title := fmt.Sprintf("Session Log: %s (%d turns)", m.session.ID[:8], len(m.logEntries))
	lines = append(lines, styles.LogTitleStyle.Render(title))
	lines = append(lines, "")

	// Calculate visible area (simple impl: show turns starting from cursor)
	startIdx := m.cursor
	endIdx := startIdx + 10 // show 10 turns at a time

	if endIdx > len(m.logEntries) {
		endIdx = len(m.logEntries)
	}

	// Render visible log entries
	for i := startIdx; i < endIdx; i++ {
		entry := m.logEntries[i]
		lines = append(lines, m.renderLogEntry(entry))
		lines = append(lines, "")
	}

	// Hint message
	if m.hasMore && endIdx >= len(m.logEntries) {
		lines = append(lines, styles.LogHintStyle.Render("(More entries available, use j/k to scroll)"))
	}

	// Join all lines
	content := strings.Join(lines, "\n")

	// Apply container style (fullscreen fill)
	container := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 2).
		Render(content)

	return container
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
