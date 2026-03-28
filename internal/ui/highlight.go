package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// HighlightMatches highlights all case-insensitive occurrences of pattern in text.
// Uses manual ANSI codes with selective reset (\x1b[22;39m) to reset only bold
// and foreground, preserving any outer background color from the row style.
func HighlightMatches(text, pattern string) string {
	if pattern == "" {
		return text
	}

	lower := strings.ToLower(text)
	lowerPattern := strings.ToLower(pattern)

	var sb strings.Builder
	lastEnd := 0
	start := 0
	for {
		idx := strings.Index(lower[start:], lowerPattern)
		if idx == -1 {
			sb.WriteString(text[lastEnd:])
			break
		}

		absIdx := start + idx
		sb.WriteString(text[lastEnd:absIdx])
		sb.WriteString("\x1b[1;38;5;220m") // bold, fg=yellow(220)
		sb.WriteString(text[absIdx : absIdx+len(pattern)])
		sb.WriteString("\x1b[22;39m") // reset bold + fg only, preserve bg
		lastEnd = absIdx + len(pattern)
		start = lastEnd
	}

	return sb.String()
}

// highlightCell renders a table cell with search highlighting.
// Strategy: render the cell with row style FIRST (correct width), then
// insert highlight ANSI codes into the rendered output. This avoids
// lipgloss measuring width on text with highlight ANSI codes, which
// caused the outer Width(contentWidth).Render() to wrap/truncate rows.
func highlightCell(style lipgloss.Style, text string, width int, pattern string) string {
	// Pad text to desired width (plain text, no ANSI codes)
	visualW := lipgloss.Width(text)
	if visualW < width {
		text += strings.Repeat(" ", width-visualW)
	}

	// Render with row style — lipgloss handles width and styling correctly
	rendered := style.Width(width).Render(text)

	if pattern == "" {
		return rendered
	}

	// Find match position in the original plain text
	lower := strings.ToLower(text)
	lowerPattern := strings.ToLower(pattern)
	idx := strings.Index(lower, lowerPattern)
	if idx == -1 {
		return rendered
	}

	// The match text is the same in both original and rendered output.
	// Search for it in the rendered string to get the byte position.
	matchText := text[idx : idx+len(lowerPattern)]
	renderedIdx := strings.Index(rendered, matchText)
	if renderedIdx == -1 {
		return rendered
	}

	// Insert highlight ANSI codes around the match in the rendered output.
	// This happens AFTER lipgloss rendering, so width calculation is unaffected.
	return rendered[:renderedIdx] +
		"\x1b[1;38;5;220m" +
		rendered[renderedIdx:renderedIdx+len(matchText)] +
		"\x1b[22;39m" +
		rendered[renderedIdx+len(matchText):]
}
