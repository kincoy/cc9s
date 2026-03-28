package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// renderSessionTable renders the session table (Approach A: manually drawn borders, title embedded in top border)
func renderSessionTable(sessions []claudefs.GlobalSession, cursor, width, height int, selectedRows map[int]struct{}, showProjectColumn bool, sortBy SessionSortField, sortAsc bool, contextLabel string, showCleanupHints bool) string {
	if len(sessions) == 0 {
		return ""
	}

	var sb strings.Builder

	// === Top border + title ===
	resourceType := styles.HighlightStyle.Render("Sessions")

	contextPart := ""
	if contextLabel != "" {
		contextPart = lipgloss.NewStyle().
			Foreground(styles.ColorPurple).
			Bold(true).
			Render(fmt.Sprintf("(%s)", contextLabel))
	}

	countStr := lipgloss.NewStyle().
		Foreground(styles.ColorWarning).
		Bold(true).
		Render(fmt.Sprintf("(%d)", len(sessions)))

	title := resourceType + contextPart + countStr
	sb.WriteString(renderTopBorder(width, title))

	// === Content area ===
	contentWidth := width - 4

	// 80x24 adaptive - hide EVENTS and SIZE columns on narrow screens
	showEventsAndSize := width >= 100

	// Column width allocation (SUMMARY is the flexible column)
	projectWidth := 0
	if showProjectColumn {
		projectWidth = 20
	}

	idWidth := 12
	statusWidth := 13
	recommendWidth := 10
	lastActiveWidth := 14
	eventsWidth := 8
	sizeWidth := 10

	colCount := 4 // ID, SUMMARY, STATUS, LAST ACTIVE
	if showProjectColumn {
		colCount++
	}
	if showCleanupHints {
		colCount++
	}
	if showEventsAndSize {
		colCount += 2
	}
	sepCount := colCount - 1

	fixedWidth := projectWidth + idWidth + statusWidth + lastActiveWidth
	if showCleanupHints {
		fixedWidth += recommendWidth
	}
	if showEventsAndSize {
		fixedWidth += eventsWidth + sizeWidth
	}

	summaryWidth := contentWidth - fixedWidth - (sepCount * 2)
	if summaryWidth < 10 {
		summaryWidth = 10
	}

	// === Header row ===
	arrow := sortArrow(sortAsc)

	type headerCol struct {
		text     string
		width    int
		align    lipgloss.Position
		sort     SessionSortField
		editable bool
	}

	allCols := []headerCol{
		{"PROJECT", projectWidth, lipgloss.Left, -1, showProjectColumn},
		{"SESSION ID", idWidth, lipgloss.Left, SortBySessionID, true},
		{"SUMMARY", summaryWidth, lipgloss.Left, -1, true},
		{"STATUS", statusWidth, lipgloss.Left, -1, true},
		{"RECOMMEND", recommendWidth, lipgloss.Left, -1, showCleanupHints},
		{"LAST ACTIVE", lastActiveWidth, lipgloss.Right, SortBySessionActivity, true},
		{"EVENTS", eventsWidth, lipgloss.Right, SortByEventCount, showEventsAndSize},
		{"SIZE", sizeWidth, lipgloss.Right, SortBySessionSize, showEventsAndSize},
	}

	var headerParts []string
	for _, col := range allCols {
		if !col.editable {
			continue
		}
		headerParts = append(headerParts, headerSep())
		isSorted := sortBy == col.sort
		text := col.text
		if isSorted {
			text += arrow
		}
		headerParts = append(headerParts, headerStyle(isSorted).Width(col.width).Align(col.align).Render(text))
	}
	headerParts = headerParts[1:] // remove leading sep

	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, headerParts...)
	sb.WriteString(renderRowBorder(headerRow))

	// === Data rows ===
	visibleHeight := height - 3
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	startIdx, endIdx := calculateScrollWindow(len(sessions), cursor, visibleHeight)

	for i := startIdx; i < endIdx; i++ {
		gs := sessions[i]

		// Row style
		var rowStyle lipgloss.Style
		_, isSelected := selectedRows[i]
		if i == cursor {
			rowStyle = styles.SelectedRowStyle
		} else if isSelected {
			rowStyle = styles.MultiSelectedStyle
		} else {
			rowStyle = styles.TableCellStyle.Faint(true)
		}

		// Format session ID
		sessionID := gs.Session.ID
		if len(sessionID) > idWidth {
			sessionID = sessionID[:idWidth-3] + "..."
		}

		statusText := styles.LifecycleStatusText(gs.Session.Lifecycle.State)
		statusStyle := styles.LifecycleStatusStyle(gs.Session.Lifecycle.State)
		assessment := claudefs.QuickAssessSession(gs.Session)
		isStale := gs.Session.Lifecycle.State == claudefs.SessionLifecycleStale

		// Apply strikethrough to ID for Stale sessions
		if isStale {
			sessionID = lipgloss.NewStyle().Strikethrough(true).Render(sessionID)
		}

		// Format project name (truncate if too long for column)
		projectName := gs.ProjectName
		if showProjectColumn && utf8.RuneCountInString(projectName) > projectWidth {
			projectName = string([]rune(projectName)[:projectWidth-3]) + "..."
		}

		rowSep := rowStyle.Render("  ")

		summaryText := gs.Session.Summary
		if summaryText == "" {
			summaryText = "-"
		}
		// Apply strikethrough to summary for Stale sessions
		if isStale {
			summaryText = lipgloss.NewStyle().Strikethrough(true).Render(summaryText)
		}

		var rowParts []string
		if showProjectColumn {
			rowParts = append(rowParts,
				rowStyle.Width(projectWidth).Render(projectName), rowSep,
				rowStyle.Width(idWidth).Render(sessionID), rowSep,
				rowStyle.Width(summaryWidth).Render(truncateSummary(summaryText, summaryWidth)), rowSep,
				statusStyle.Inherit(rowStyle).Width(statusWidth).Render(statusText),
			)
		} else {
			rowParts = append(rowParts,
				rowStyle.Width(idWidth).Render(sessionID), rowSep,
				rowStyle.Width(summaryWidth).Render(truncateSummary(summaryText, summaryWidth)), rowSep,
				statusStyle.Inherit(rowStyle).Width(statusWidth).Render(statusText),
			)
		}

		if showCleanupHints {
			var recStyle lipgloss.Style
			switch assessment.Recommendation {
			case claudefs.RecommendDelete:
				recStyle = rowStyle.Width(recommendWidth).Foreground(styles.ColorError)
			case claudefs.RecommendMaybe:
				recStyle = rowStyle.Width(recommendWidth).Foreground(styles.ColorWarning)
			default:
				recStyle = rowStyle.Width(recommendWidth).Foreground(styles.ColorActive)
			}
			rowParts = append(rowParts,
				rowSep,
				recStyle.Render(string(assessment.Recommendation)),
			)
		}

		rowParts = append(rowParts,
			rowSep,
			rowStyle.Width(lastActiveWidth).Align(lipgloss.Right).Render(claudefs.FormatTimeAgo(gs.Session.LastActiveAt)),
		)

		if showEventsAndSize {
			rowParts = append(rowParts,
				rowSep,
				rowStyle.Width(eventsWidth).Align(lipgloss.Right).Render(claudefs.FormatEventCount(gs.Session.EventCount)),
				rowSep,
				rowStyle.Width(sizeWidth).Align(lipgloss.Right).Render(claudefs.FormatSize(gs.Session.FileSize)),
			)
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top, rowParts...)
		row = rowStyle.Width(contentWidth).Render(row)
		sb.WriteString(renderRowBorder(row))
	}

	// Pad empty rows
	fillEmptyRows(&sb, contentWidth, endIdx-startIdx, visibleHeight)

	// === Bottom border ===
	sb.WriteString(renderBottomBorder(width))

	return sb.String()
}

// truncateSummary truncates summary text by display width (reserves 3 chars for "...")
func truncateSummary(s string, maxLen int) string {
	if lipgloss.Width(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	var width int
	limit := maxLen - 3
	if limit < 1 {
		limit = 1
	}
	for i, r := range runes {
		rw := lipgloss.Width(string(r))
		if width+rw > limit {
			return string(runes[:i]) + "..."
		}
		width += rw
	}
	return s
}
