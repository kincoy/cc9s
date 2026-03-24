package ui

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

func renderHeader(width int, contextLabel, stats string) string {
	return renderHeaderWithFilter(width, contextLabel, stats, 0, 0)
}

// renderHeaderWithFilter renders the header with optional filtered-count display.
func renderHeaderWithFilter(width int, contextLabel, stats string, filteredCount, totalCount int) string {
	statsLabel := stats
	if totalCount > 0 && filteredCount != totalCount {
		statsLabel = fmt.Sprintf("%d/%d shown / %s", filteredCount, totalCount, stats)
	}

	if width < 100 {
		logo := styles.TitleStyle.Render("cc9s")
		statsRendered := styles.NormalStyle.Render(fmt.Sprintf("%s / %s", contextLabel, statsLabel))
		sep := styles.DimStyle.Render(" │ ")

		content := fmt.Sprintf(" %s%s%s ", logo, sep, statsRendered)

		return styles.PanelBorderStyle.
			Width(width).
			Render(content)
	}

	logo := styles.TitleStyle.Render("cc9s v0.1.3")
	statsRendered := styles.NormalStyle.Render(fmt.Sprintf("%s / %s", contextLabel, statsLabel))
	clock := styles.DimStyle.Render(time.Now().Format("15:04:05"))
	sep := styles.DimStyle.Render(" │ ")

	left := fmt.Sprintf("%s%s%s", logo, sep, statsRendered)
	right := clock

	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if gap < 1 {
		gap = 1
	}
	padding := lipgloss.NewStyle().Width(gap).Render("")

	content := fmt.Sprintf(" %s%s%s ", left, padding, right)

	return styles.PanelBorderStyle.
		Width(width).
		Render(content)
}
