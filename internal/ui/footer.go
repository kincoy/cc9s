package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// FooterMode footer mode
type FooterMode int

const (
	FooterModeNormal FooterMode = iota
	FooterModeSearch
	FooterModeCommand
)

// OverlayType overlay type
type OverlayType int

const (
	OverlayNone OverlayType = iota
	OverlayDetail
	OverlayLog
	OverlayDialog
	OverlayHelp
)

// FooterContext footer state machine context
type FooterContext struct {
	Screen        Screen
	Mode          FooterMode
	Overlay       OverlayType
	DialogIsAlert bool
	HasMulti      bool
	SelCount      int
}

// KeyHint keyboard shortcut hint
type KeyHint struct {
	Key   string
	Label string
}

// hintsForContext returns the key hints for the given FooterContext
func hintsForContext(ctx FooterContext) []KeyHint {
	// Overlay has highest priority
	switch ctx.Overlay {
	case OverlayDialog:
		if ctx.DialogIsAlert {
			return []KeyHint{
				{Key: "Any key", Label: "Close"},
			}
		}
		return []KeyHint{
			{Key: "y", Label: "Confirm"},
			{Key: "n", Label: "Cancel"},
		}

	case OverlayHelp:
		return []KeyHint{
			{Key: "Esc", Label: "Close"},
		}

	case OverlayDetail:
		if ctx.Screen == ScreenSkills || ctx.Screen == ScreenAgents {
			return []KeyHint{
				{Key: "e", Label: "Edit"},
				{Key: "Esc", Label: "Close detail"},
				{Key: "?", Label: "Help"},
			}
		}
		return []KeyHint{
			{Key: "Esc", Label: "Close detail"},
			{Key: "?", Label: "Help"},
		}

	case OverlayLog:
		return []KeyHint{
			{Key: "j/k", Label: "Scroll"},
			{Key: "g/G", Label: "Top/Bottom"},
			{Key: "Esc", Label: "Close log"},
			{Key: "?", Label: "Help"},
		}
	}

	// No overlay, check Mode
	switch ctx.Mode {
	case FooterModeSearch:
		return []KeyHint{
			{Key: "Esc", Label: "Cancel search"},
			{Key: "Enter", Label: "Confirm"},
		}

	case FooterModeCommand:
		return []KeyHint{
			{Key: "Esc", Label: "Cancel command"},
			{Key: "Enter", Label: "Execute"},
			{Key: "Tab", Label: "Complete"},
		}
	}

	// Normal mode, by Screen
	switch ctx.Screen {
	case ScreenProjects:
		return []KeyHint{
			{Key: "q", Label: "Quit"},
			{Key: "j/k", Label: "Navigate"},
			{Key: "s/S", Label: "Sort"},
			{Key: "Enter", Label: "Open"},
			{Key: "d", Label: "Detail"},
			{Key: "/", Label: "Search"},
			{Key: ":", Label: "Cmd"},
			{Key: "?", Label: "Help"},
		}

	case ScreenSessions:
		hints := []KeyHint{
			{Key: "q", Label: "Quit"},
			{Key: "j/k", Label: "Navigate"},
			{Key: "s/S", Label: "Sort"},
			{Key: "Enter", Label: "Resume"},
			{Key: "d", Label: "Detail"},
			{Key: "l", Label: "Logs"},
			{Key: "Space", Label: "Select"},
			{Key: "/", Label: "Search"},
			{Key: ":", Label: "Cmd"},
			{Key: "Esc", Label: "Back"},
		}

		// Append multi-select hint
		if ctx.HasMulti {
			hints = append(hints, KeyHint{Key: "Ctrl+D", Label: "Delete"})
		}

		hints = append(hints, KeyHint{Key: "0", Label: "All ctx"})
		hints = append(hints, KeyHint{Key: "?", Label: "Help"})

		return hints

	case ScreenSkills:
		return []KeyHint{
			{Key: "q", Label: "Quit"},
			{Key: "j/k", Label: "Navigate"},
			{Key: "s/S", Label: "Sort"},
			{Key: "d", Label: "Detail"},
			{Key: "e", Label: "Edit"},
			{Key: "/", Label: "Search"},
			{Key: ":", Label: "Cmd"},
			{Key: "0", Label: "All ctx"},
			{Key: "?", Label: "Help"},
		}

	case ScreenAgents:
		return []KeyHint{
			{Key: "q", Label: "Quit"},
			{Key: "j/k", Label: "Navigate"},
			{Key: "s/S", Label: "Sort"},
			{Key: "d", Label: "Detail"},
			{Key: "e", Label: "Edit"},
			{Key: "/", Label: "Search"},
			{Key: ":", Label: "Cmd"},
			{Key: "0", Label: "All ctx"},
			{Key: "?", Label: "Help"},
		}
	}

	// Default
	return []KeyHint{
		{Key: "q", Label: "Quit"},
	}
}

// renderFooterWithHints renders the footer with a list of key hints
func renderFooterWithHints(width int, hints []KeyHint) string {
	const separator = "  "

	var rendered []string
	for _, h := range hints {
		key := styles.FooterKeyStyle.Render("<" + h.Key + ">")
		label := styles.NormalStyle.Render(h.Label)
		rendered = append(rendered, key+" "+label)
	}

	content := " " + strings.Join(rendered, separator)

	// Auto-truncate if content exceeds width
	if lipgloss.Width(content) > width {
		// Simplified: keep only the first N hints until within width
		rendered = rendered[:len(rendered)-1]
		for len(rendered) > 4 && lipgloss.Width(" "+strings.Join(rendered, separator)) > width {
			rendered = rendered[:len(rendered)-1]
		}
		content = " " + strings.Join(rendered, separator)
	}

	return styles.FooterStyle.Width(width).Render(content)
}

// renderFlashFooter renders flash message in the footer
func renderFlashFooter(width int, message string, isError bool) string {
	if isError {
		return lipgloss.NewStyle().
			Background(styles.ColorError).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Width(width).
			Render(" " + message)
	}
	return lipgloss.NewStyle().
		Background(styles.ColorFlashSuccess).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Width(width).
		Render(" " + message)
}
