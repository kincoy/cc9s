package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

type ResourceHelpSection struct {
	Title string
	Lines []KeyHint
}

func buildHelpLines(registry *ResourceRegistry, active ResourceDescriptor) []string {
	title := styles.HeaderStyle.Render("Keyboard Shortcuts")
	divider := styles.BreadcrumbStyle.Render(strings.Repeat("─", 20))
	currentResource := styles.FooterStyle.Render("Current resource: " + active.DisplayName)

	lines := []string{
		"",
		"  " + title,
		"  " + divider,
		"  " + currentResource,
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
		"  " + styles.FooterKeyStyle.Render("Ctrl+D") + styles.FooterStyle.Render("   Scroll half-page down"),
		"  " + styles.FooterKeyStyle.Render("Ctrl+U") + styles.FooterStyle.Render("   Scroll half-page up"),
		"  " + styles.FooterKeyStyle.Render("PgDn") + styles.FooterStyle.Render("     Scroll page down"),
		"  " + styles.FooterKeyStyle.Render("PgUp") + styles.FooterStyle.Render("     Scroll page up"),
		"",
		"  " + styles.HeaderStyle.Render("Sorting"),
		"  " + styles.FooterKeyStyle.Render("s") + styles.FooterStyle.Render("         Cycle sort field"),
		"  " + styles.FooterKeyStyle.Render("S") + styles.FooterStyle.Render("         Reverse sort order"),
	}
	for _, descriptor := range registry.ordered {
		if descriptor.HelpSection == nil {
			continue
		}
		section := descriptor.HelpSection()
		if section.Title == "" || len(section.Lines) == 0 {
			continue
		}
		lines = append(lines, "", "  "+styles.HeaderStyle.Render(section.Title))
		lines = append(lines, renderKeyHintHelpLines(section.Lines)...)
	}
	lines = append(lines, "", "  "+styles.HeaderStyle.Render("Context"))
	lines = append(lines,
		"  "+styles.FooterKeyStyle.Render(":context")+styles.FooterStyle.Render("   Switch context (all / project name)"),
	)
	for _, descriptor := range registry.ordered {
		lines = append(lines,
			"  "+styles.FooterKeyStyle.Render(":"+descriptor.CommandName)+styles.FooterStyle.Render("    Open "+strings.ToLower(descriptor.DisplayName)+" resource"),
		)
	}
	lines = append(lines,
		"  "+styles.FooterKeyStyle.Render("0")+styles.FooterStyle.Render("         Switch to all projects"),
		"  "+styles.FooterKeyStyle.Render("Tab")+styles.FooterStyle.Render("       Auto-complete commands"),
		"",
		"  "+styles.HeaderStyle.Render("Dialog"),
		"  "+styles.FooterKeyStyle.Render("y")+styles.FooterStyle.Render("         Confirm"),
		"  "+styles.FooterKeyStyle.Render("n")+styles.FooterStyle.Render("         Cancel"),
		"",
	)

	return lines
}

func renderHelp(width, height int, registry *ResourceRegistry, active ResourceDescriptor, scroll int) string {
	lines := buildHelpLines(registry, active)

	if scroll < 0 {
		scroll = 0
	}
	maxScroll := len(lines) - height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}

	end := scroll + height
	if end > len(lines) {
		end = len(lines)
	}

	content := strings.Join(lines[scroll:end], "\n")
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(styles.ColorHighlight).
		Padding(0, 1).
		Render(content)
}

type helpLine struct {
	key     string
	label   string
	enabled bool
}

func renderHelpLines(lines ...helpLine) []string {
	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		if !line.enabled {
			continue
		}
		rendered = append(rendered, "  "+styles.FooterKeyStyle.Render(line.key)+styles.FooterStyle.Render(line.label))
	}
	return rendered
}

func renderKeyHintHelpLines(hints []KeyHint) []string {
	rendered := make([]string, 0, len(hints))
	for _, hint := range hints {
		rendered = append(rendered, "  "+styles.FooterKeyStyle.Render(hint.Key)+styles.FooterStyle.Render(hint.Label))
	}
	return rendered
}
