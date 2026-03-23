package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

func renderHelp(width, height int) string {
	title := styles.HeaderStyle.Render("Keyboard Shortcuts")
	divider := styles.BreadcrumbStyle.Render(strings.Repeat("─", 20))

	lines := []string{
		"",
		"  " + title,
		"  " + divider,
		"",
		"  " + styles.HeaderStyle.Render("General"),
		"  " + styles.FooterKeyStyle.Render("q") + styles.FooterStyle.Render("         Quit cc9s"),
		"  " + styles.FooterKeyStyle.Render("?") + styles.FooterStyle.Render("         Toggle this help"),
		"",
		"  " + styles.HeaderStyle.Render("Navigation"),
		"  " + styles.FooterKeyStyle.Render("j / ↓") + styles.FooterStyle.Render("     Move down"),
		"  " + styles.FooterKeyStyle.Render("k / ↑") + styles.FooterStyle.Render("     Move up"),
		"  " + styles.FooterKeyStyle.Render("g") + styles.FooterStyle.Render("         Go to top"),
		"  " + styles.FooterKeyStyle.Render("G") + styles.FooterStyle.Render("         Go to bottom"),
		"  " + styles.FooterKeyStyle.Render("Enter") + styles.FooterStyle.Render("     Open project / Resume session"),
		"  " + styles.FooterKeyStyle.Render("Esc") + styles.FooterStyle.Render("       Go back / Cancel"),
		"",
		"  " + styles.HeaderStyle.Render("Sorting"),
		"  " + styles.FooterKeyStyle.Render("s") + styles.FooterStyle.Render("         Cycle sort field"),
		"  " + styles.FooterKeyStyle.Render("S") + styles.FooterStyle.Render("         Reverse sort order"),
		"",
		"  " + styles.HeaderStyle.Render("Session Operations"),
		"  " + styles.FooterKeyStyle.Render("d") + styles.FooterStyle.Render("         View session details"),
		"  " + styles.FooterKeyStyle.Render("l") + styles.FooterStyle.Render("         View session log"),
		"  " + styles.FooterKeyStyle.Render("/") + styles.FooterStyle.Render("         Search sessions"),
		"",
		"  " + styles.HeaderStyle.Render("Multi-select & Delete"),
		"  " + styles.FooterKeyStyle.Render("Space") + styles.FooterStyle.Render("      Toggle select session"),
		"  " + styles.FooterKeyStyle.Render("Ctrl+D") + styles.FooterStyle.Render("     Delete selected session(s)"),
		"",
		"  " + styles.HeaderStyle.Render("Context"),
		"  " + styles.FooterKeyStyle.Render("0") + styles.FooterStyle.Render("         Switch to all projects"),
		"  " + styles.FooterKeyStyle.Render(":context") + styles.FooterStyle.Render("   Switch context (all / project name)"),
		"  " + styles.FooterKeyStyle.Render("Tab") + styles.FooterStyle.Render("       Auto-complete commands"),
		"",
		"  " + styles.HeaderStyle.Render("Dialog"),
		"  " + styles.FooterKeyStyle.Render("y") + styles.FooterStyle.Render("         Confirm"),
		"  " + styles.FooterKeyStyle.Render("n") + styles.FooterStyle.Render("         Cancel"),
		"",
	}

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Render(content)
}
