package ui

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/kincoy/cc9s/internal/ui/styles"
	"github.com/kincoy/cc9s/internal/version"
)

func renderHeader(width int, resourceLabel, contextLabel, stats string, currentTime time.Time) string {
	return renderHeaderWithFilter(width, resourceLabel, contextLabel, stats, 0, 0, currentTime)
}

// renderHeaderWithFilter renders the header with optional filtered-count display.
func renderHeaderWithFilter(width int, resourceLabel, contextLabel, stats string, filteredCount, totalCount int, currentTime time.Time) string {
	statsLabel := stats
	if totalCount > 0 && filteredCount != totalCount {
		statsLabel = fmt.Sprintf("%d/%d shown / %s", filteredCount, totalCount, stats)
	}

	scopeLabel := resourceLabel
	if contextLabel != "" {
		scopeLabel = fmt.Sprintf("%s: %s", resourceLabel, contextLabel)
	}

	if width < 100 {
		logo := styles.TitleStyle.Render("cc9s")
		statsRendered := styles.NormalStyle.Render(fmt.Sprintf("%s / %s", scopeLabel, statsLabel))
		sep := styles.DimStyle.Render(" │ ")

		content := fmt.Sprintf(" %s%s%s ", logo, sep, statsRendered)

		return styles.PanelBorderStyle.
			Width(width).
			Render(content)
	}

	logo := styles.TitleStyle.Render("cc9s v" + version.Version)
	statsRendered := styles.NormalStyle.Render(fmt.Sprintf("%s / %s", scopeLabel, statsLabel))
	clock := styles.DimStyle.Render(currentTime.Format("15:04:05"))
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
