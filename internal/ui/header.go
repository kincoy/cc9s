package ui

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

func renderHeader(width int, contextLabel string, totalSessions, activeCount int) string {
	return renderHeaderWithFilter(width, contextLabel, totalSessions, activeCount, 0, 0)
}

// renderHeaderWithFilter renders the header with filter count display
func renderHeaderWithFilter(width int, contextLabel string, totalSessions, activeCount, filteredCount, totalCount int) string {
	if width < 100 {
		// Narrow screen: show only core stats
		logo := styles.TitleStyle.Render("cc9s")
		var stats string
		if totalCount > 0 && filteredCount != totalCount {
			stats = fmt.Sprintf("%s / %d/%d sessions / %d active", contextLabel, filteredCount, totalCount, activeCount)
		} else {
			stats = fmt.Sprintf("%s / %d sessions / %d active", contextLabel, totalSessions, activeCount)
		}
		statsRendered := styles.NormalStyle.Render(stats)
		sep := styles.DimStyle.Render(" │ ")

		content := fmt.Sprintf(" %s%s%s ", logo, sep, statsRendered)

		return styles.PanelBorderStyle.
			Width(width).
			Render(content)
	}

	logo := styles.TitleStyle.Render("cc9s v0.1.0")

	var stats string
	if totalCount > 0 && filteredCount != totalCount {
		stats = fmt.Sprintf("%s / %d/%d sessions / %d active", contextLabel, filteredCount, totalCount, activeCount)
	} else {
		stats = fmt.Sprintf("%s / %d sessions / %d active", contextLabel, totalSessions, activeCount)
	}
	statsRendered := styles.NormalStyle.Render(stats)

	clock := styles.DimStyle.Render(time.Now().Format("15:04:05"))

	sep := styles.DimStyle.Render(" │ ")

	left := fmt.Sprintf("%s%s%s", logo, sep, statsRendered)
	right := clock

	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 4 // subtract left/right padding
	if gap < 1 {
		gap = 1
	}
	padding := lipgloss.NewStyle().Width(gap).Render("")

	content := fmt.Sprintf(" %s%s%s ", left, padding, right)

	return styles.PanelBorderStyle.
		Width(width).
		Render(content)
}
