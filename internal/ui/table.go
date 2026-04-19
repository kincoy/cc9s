package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// renderProjectTable renders the project table (Approach A: manually drawn borders, title embedded in top border)
func renderProjectTable(projects []claudefs.Project, cursor, width, height int, sortBy SortField, sortAsc bool, showHealthColumn bool, showCleanupColumn bool, projectHealth map[string]int, searchPattern string) string {
	if len(projects) == 0 {
		return renderEmptyState(width, height)
	}

	var sb strings.Builder

	// === Top border + title ===
	resourceType := styles.HighlightStyle.Render("Projects")
	countStr := lipgloss.NewStyle().
		Foreground(styles.ColorWarning).
		Bold(true).
		Render(fmt.Sprintf("(%d)", len(projects)))
	title := resourceType + countStr

	sb.WriteString(renderTopBorder(width, title))

	// === Content area ===
	contentWidth := width - 4

	// 80x24 adaptive - hide SIZE column on narrow screens
	showSize := width >= 100
	showPath := width >= 140

	pathWidth := 0
	sessionsWidth := 10
	skillsWidth := 10
	agentsWidth := 10
	lastActiveWidth := 14
	sizeWidth := 10
	healthWidth := 0
	if showHealthColumn {
		healthWidth = 8
	}
	cleanupWidth := 0
	if showCleanupColumn {
		cleanupWidth = 10
	}

	sepCount := 5
	if showPath {
		sepCount++
	}
	if !showSize {
		sepCount--
	}
	if showHealthColumn {
		sepCount++
	}
	if showCleanupColumn {
		sepCount++
	}
	sepWidth := sepCount * 2

	fixedWidth := sessionsWidth + skillsWidth + agentsWidth + lastActiveWidth + healthWidth + cleanupWidth + sepWidth
	if showSize {
		fixedWidth += sizeWidth
	}
	remainingWidth := contentWidth - fixedWidth
	nameWidth := remainingWidth
	if showPath {
		pathWidth = remainingWidth * 2 / 3
		if pathWidth < 28 {
			pathWidth = 28
		}
		nameWidth = remainingWidth - pathWidth
	}
	if nameWidth < 10 {
		nameWidth = 10
	}
	if showPath && pathWidth < 16 {
		pathWidth = 16
	}

	// === Header row ===
	headers := []struct {
		text  string
		width int
		align lipgloss.Position
		field SortField
	}{
		{"NAME", nameWidth, lipgloss.Left, SortByName},
		{"PATH", pathWidth, lipgloss.Left, SortByPath},
		{"SESSIONS", sessionsWidth, lipgloss.Right, SortBySessionCount},
		{"LOCAL SK", skillsWidth, lipgloss.Right, SortBySkillCount},
		{"LOCAL AG", agentsWidth, lipgloss.Right, SortByAgentCount},
		{"LAST ACTIVE", lastActiveWidth, lipgloss.Right, SortByActivity},
		{"HEALTH", healthWidth, lipgloss.Right, SortByHealth},
		{"STATUS", cleanupWidth, lipgloss.Left, -1},
		{"SIZE", sizeWidth, lipgloss.Right, SortBySize},
	}
	visibleHeaders := make([]struct {
		text  string
		width int
		align lipgloss.Position
		field SortField
	}, 0, len(headers))
	for _, h := range headers {
		if h.width <= 0 {
			continue
		}
		if h.field == SortBySize && !showSize {
			continue
		}
		if h.text == "HEALTH" && !showHealthColumn {
			continue
		}
		if h.text == "STATUS" && !showCleanupColumn {
			continue
		}
		visibleHeaders = append(visibleHeaders, h)
	}

	var headerParts []string
	for i, h := range visibleHeaders {
		if i > 0 {
			headerParts = append(headerParts, headerSep())
		}
		if h.field == sortBy {
			headerParts = append(headerParts, headerStyle(true).Width(h.width).Align(h.align).Render(renderProjectHeaderLabel(h.text, h.width, sortAsc)))
		} else {
			headerParts = append(headerParts, headerStyle(false).Width(h.width).Align(h.align).Render(h.text))
		}
	}

	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, headerParts...)
	sb.WriteString(renderRowBorder(headerRow))

	// === Data rows ===
	visibleHeight := height - 3
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	startIdx, endIdx := calculateScrollWindow(len(projects), cursor, visibleHeight)

	for i := startIdx; i < endIdx; i++ {
		project := projects[i]

		// Truncate long project names (UTF-8 safe)
		name := project.Name
		if utf8.RuneCountInString(name) > nameWidth-3 {
			name = string([]rune(name)[:nameWidth-3]) + "..."
		}
		// Apply search highlighting after truncation, before rowStyle render
		// Use highlightCell to pad first, then highlight, then render - avoids
		// Width().Render() miscalculating when ANSI escape codes are present.

		path := ""
		if showPath {
			path = truncateProjectPath(project.Path, pathWidth)
		}

		sessions := fmt.Sprintf("%d", project.SessionCount)
		skills := fmt.Sprintf("%d", projectLocalSkillTotal(project))
		agents := fmt.Sprintf("%d", project.AgentCount)
		lastActive := claudefs.FormatTimeAgo(project.LastActiveAt)

		var rowStyle lipgloss.Style
		if i == cursor {
			rowStyle = styles.SelectedRowStyle
		} else {
			rowStyle = styles.TableCellStyle.Faint(true)
		}
		rowSep := rowStyle.Render("  ")

		var rowParts []string
		rowParts = append(rowParts,
			highlightCell(rowStyle, name, nameWidth, searchPattern),
			rowSep,
		)
		if showPath {
			rowParts = append(rowParts,
				highlightCell(rowStyle, path, pathWidth, searchPattern),
				rowSep,
			)
		}
		rowParts = append(rowParts,
			rowStyle.Width(sessionsWidth).Align(lipgloss.Right).Render(sessions),
			rowSep,
			rowStyle.Width(skillsWidth).Align(lipgloss.Right).Render(skills),
			rowSep,
			rowStyle.Width(agentsWidth).Align(lipgloss.Right).Render(agents),
			rowSep,
			rowStyle.Width(lastActiveWidth).Align(lipgloss.Right).Render(lastActive),
		)
		if showHealthColumn {
			health := projectHealth[project.Name]
			healthColor := styles.ColorNormal
			if health >= 70 {
				healthColor = styles.ColorActive
			} else if health >= 50 {
				healthColor = styles.ColorWarning
			} else {
				healthColor = styles.ColorError
			}
			healthField := lipgloss.NewStyle().
				Foreground(healthColor).
				Inherit(rowStyle).
				Width(healthWidth).
				Align(lipgloss.Right).
				Render(fmt.Sprintf("%d", health))
			rowParts = append(rowParts,
				rowSep,
				healthField,
			)
		}
		if showCleanupColumn {
			statusText := "OK"
			statusColor := styles.ColorActive
			if !project.PathExists {
				statusText = "Deleted"
				statusColor = styles.ColorError
			}
			statusField := lipgloss.NewStyle().
				Foreground(statusColor).
				Inherit(rowStyle).
				Width(cleanupWidth).
				Align(lipgloss.Left).
				Render(statusText)
			rowParts = append(rowParts,
				rowSep,
				statusField,
			)
		}
		if showSize {
			size := claudefs.FormatSize(project.TotalSize)
			rowParts = append(rowParts,
				rowSep,
				rowStyle.Width(sizeWidth).Align(lipgloss.Right).Render(size),
			)
		}

		rowContent := lipgloss.JoinHorizontal(lipgloss.Top, rowParts...)
		row := rowStyle.Width(contentWidth).Render(rowContent)

		sb.WriteString(renderRowBorder(row))
	}

	// Pad empty rows
	fillEmptyRows(&sb, contentWidth, endIdx-startIdx, visibleHeight)

	// === Bottom border ===
	sb.WriteString(renderBottomBorder(width))

	return sb.String()
}

// renderEmptyState renders the empty state placeholder
func renderEmptyState(width, height int) string {
	msg := "No projects found in ~/.claude/projects/"
	hint := "Create a project with Claude Code to get started."

	content := lipgloss.JoinVertical(lipgloss.Center,
		styles.TableCellStyle.Foreground(styles.ColorDim).Render(msg),
		styles.TableCellStyle.Foreground(styles.ColorDim).Render(hint),
	)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

// calculateScrollWindow calculates the scroll window start/end indices
func calculateScrollWindow(totalItems, cursor, visibleHeight int) (start, end int) {
	if totalItems <= visibleHeight {
		return 0, totalItems
	}

	halfHeight := visibleHeight / 2
	start = cursor - halfHeight
	end = start + visibleHeight

	if start < 0 {
		start = 0
		end = visibleHeight
	}
	if end > totalItems {
		end = totalItems
		start = totalItems - visibleHeight
	}

	return start, end
}

// renderCenteredText renders text centered within the given area
func renderCenteredText(text string, width, height int) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
		styles.TableCellStyle.Foreground(styles.ColorDim).Render(text))
}

func renderProjectHeaderLabel(label string, width int, sortAsc bool) string {
	arrow := sortArrow(sortAsc)
	if lipgloss.Width(label)+lipgloss.Width(arrow) <= width {
		return label + arrow
	}

	runes := []rune(label)
	for len(runes) > 0 && lipgloss.Width(string(runes))+lipgloss.Width(arrow) > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + arrow
}

func truncateProjectPath(path string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(path) <= width {
		return path
	}
	runes := []rune(path)
	if width <= 1 {
		return string(runes[:width])
	}
	return "…" + string(runes[len(runes)-(width-1):])
}
