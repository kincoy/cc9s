package styles

import (
	"charm.land/lipgloss/v2"
	"github.com/kincoy/cc9s/internal/claudefs"
)

var (
	ActiveStyle = lipgloss.NewStyle().
			Foreground(ColorActive).
			Bold(true)

	IdleStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	CompletedStyle = lipgloss.NewStyle().
			Foreground(ColorCompleted).
			Bold(true)

	StaleStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)
)

func LifecycleStatusStyle(state claudefs.SessionLifecycleState) lipgloss.Style {
	switch state {
	case claudefs.SessionLifecycleActive:
		return ActiveStatusStyle
	case claudefs.SessionLifecycleIdle:
		return IdleStatusStyle
	case claudefs.SessionLifecycleStale:
		return StaleStatusStyle
	default:
		return CompletedStatusStyle
	}
}

func LifecycleStatusText(state claudefs.SessionLifecycleState) string {
	switch state {
	case claudefs.SessionLifecycleActive:
		return "● Active"
	case claudefs.SessionLifecycleIdle:
		return "◐ Idle"
	case claudefs.SessionLifecycleStale:
		return "▲ Stale"
	default:
		return "○ Completed"
	}
}
