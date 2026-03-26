package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// ThickBorder character set
const (
	borderTopLeft     = "┏"
	borderTopRight    = "┓"
	borderBottomLeft  = "┗"
	borderBottomRight = "┛"
	borderHorizontal  = "━"
	borderVertical    = "┃"
)

// borderColor returns the border rendering style
func borderColor() lipgloss.Style {
	return styles.DimStyle
}

// renderTopBorder builds the top border with an embedded title
// Format: ┏━ title ━━━━━━━━━━━┓
func renderTopBorder(width int, title string) string {
	bc := borderColor()
	leftPart := bc.Render(borderTopLeft+borderHorizontal) + " " + title + " "
	leftWidth := lipgloss.Width(leftPart)

	fillLen := width - leftWidth - 1
	if fillLen < 0 {
		fillLen = 0
	}

	return leftPart + bc.Render(strings.Repeat(borderHorizontal, fillLen)+borderTopRight) + "\n"
}

// renderBottomBorder builds the bottom border
func renderBottomBorder(width int) string {
	bc := borderColor()
	return bc.Render(borderBottomLeft + strings.Repeat(borderHorizontal, width-2) + borderBottomRight)
}

// renderRowBorder wraps a row of content with border vertical lines
func renderRowBorder(content string) string {
	bc := borderColor()
	return bc.Render(borderVertical) + " " + content + " " + bc.Render(borderVertical) + "\n"
}

// renderEmptyRow fills an empty row (keeps bottom border in place)
func renderEmptyRow(contentWidth int) string {
	bc := borderColor()
	emptyContent := lipgloss.NewStyle().Width(contentWidth).Render("")
	return bc.Render(borderVertical) + " " + emptyContent + " " + bc.Render(borderVertical) + "\n"
}

// fillEmptyRows pads empty rows up to visibleHeight
func fillEmptyRows(sb *strings.Builder, contentWidth, actualRows, visibleHeight int) {
	for i := actualRows; i < visibleHeight; i++ {
		sb.WriteString(renderEmptyRow(contentWidth))
	}
}

// headerStyle returns the header cell style
func headerStyle(sorted bool) lipgloss.Style {
	if sorted {
		return lipgloss.NewStyle().
			Foreground(styles.ColorHighlight).
			Background(styles.ColorTableHeaderBg).
			Bold(true)
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(styles.ColorTableHeaderBg).
		Bold(true)
}

// headerSep returns the header column separator (with background color for continuity).
// It is a function instead of a var so it reads the color at call time (after theme is applied).
func headerSep() string {
	return lipgloss.NewStyle().
		Background(styles.ColorTableHeaderBg).
		Render("  ")
}

// sortArrow returns the sort direction arrow
func sortArrow(asc bool) string {
	if asc {
		return " ↑"
	}
	return " ↓"
}
