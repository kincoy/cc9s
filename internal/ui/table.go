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
func renderProjectTable(projects []claudefs.Project, cursor, width, height int, sortBy SortField, sortAsc bool) string {
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

	sessionsWidth := 10
	lastActiveWidth := 14
	sizeWidth := 10

	sepCount := 3
	if !showSize {
		sepCount = 2
	}
	sepWidth := sepCount * 2

	nameWidth := contentWidth - sessionsWidth - lastActiveWidth - sepWidth
	if showSize {
		nameWidth -= sizeWidth
	}
	if nameWidth < 10 {
		nameWidth = 10
	}

	// === Header row ===
	arrow := sortArrow(sortAsc)

	sortedIdx := -1
	switch sortBy {
	case SortByName:
		sortedIdx = 0
	case SortBySessionCount:
		sortedIdx = 1
	case SortByActivity:
		sortedIdx = 2
	case SortBySize:
		sortedIdx = 3
	}

	headers := []struct {
		text   string
		width  int
		align  lipgloss.Position
		field  SortField
	}{
		{"NAME", nameWidth, lipgloss.Left, SortByName},
		{"SESSIONS", sessionsWidth, lipgloss.Right, SortBySessionCount},
		{"LAST ACTIVE", lastActiveWidth, lipgloss.Right, SortByActivity},
		{"SIZE", sizeWidth, lipgloss.Right, SortBySize},
	}

	var headerParts []string
	for i, h := range headers {
		if i > 0 {
			headerParts = append(headerParts, headerSep)
		}
		if i == sortedIdx {
			headerParts = append(headerParts, headerStyle(true).Width(h.width).Align(h.align).Render(h.text+arrow))
		} else {
			headerParts = append(headerParts, headerStyle(false).Width(h.width).Align(h.align).Render(h.text))
		}
	}

	// Remove unwanted columns
	if !showSize {
		headerParts = headerParts[:len(headerParts)-2] // remove trailing sep + SIZE
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

		sessions := fmt.Sprintf("%d", project.SessionCount)
		lastActive := claudefs.FormatTimeAgo(project.LastActiveAt)

		var rowParts []string
		rowParts = append(rowParts,
			lipgloss.NewStyle().Width(nameWidth).Render(name),
			lipgloss.NewStyle().Render("  "),
			lipgloss.NewStyle().Width(sessionsWidth).Align(lipgloss.Right).Render(sessions),
			lipgloss.NewStyle().Render("  "),
			lipgloss.NewStyle().Width(lastActiveWidth).Align(lipgloss.Right).Render(lastActive),
		)
		if showSize {
			size := claudefs.FormatSize(project.TotalSize)
			rowParts = append(rowParts,
				lipgloss.NewStyle().Render("  "),
				lipgloss.NewStyle().Width(sizeWidth).Align(lipgloss.Right).Render(size),
			)
		}

		rowContent := lipgloss.JoinHorizontal(lipgloss.Top, rowParts...)

		var row string
		if i == cursor {
			row = styles.SelectedRowStyle.Width(contentWidth).Render(rowContent)
		} else {
			row = styles.TableCellStyle.Width(contentWidth).Render(rowContent)
		}

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
