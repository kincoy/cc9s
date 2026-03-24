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

func SkillStatusStyle(status claudefs.SkillStatus) lipgloss.Style {
	switch status {
	case claudefs.SkillStatusInvalid:
		return InvalidSkillStatusStyle
	default:
		return ReadySkillStatusStyle
	}
}

func SkillStatusText(status claudefs.SkillStatus) string {
	switch status {
	case claudefs.SkillStatusInvalid:
		return "! Invalid"
	default:
		return "✓ Ready"
	}
}

func AgentStatusStyle(status claudefs.AgentStatus) lipgloss.Style {
	switch status {
	case claudefs.AgentStatusInvalid:
		return InvalidAgentStatusStyle
	default:
		return ReadyAgentStatusStyle
	}
}

func AgentStatusText(status claudefs.AgentStatus) string {
	switch status {
	case claudefs.AgentStatusInvalid:
		return "! Invalid"
	default:
		return "✓ Ready"
	}
}

func SkillScopeStyle(scope claudefs.SkillScope) lipgloss.Style {
	switch scope {
	case claudefs.SkillScopeProject:
		return ProjectSkillSourceStyle
	case claudefs.SkillScopePlugin:
		return PluginSkillSourceStyle
	default:
		return UserSkillSourceStyle
	}
}

func SkillScopeText(scope claudefs.SkillScope) string {
	switch scope {
	case claudefs.SkillScopeProject:
		return "Project"
	case claudefs.SkillScopePlugin:
		return "Plugin"
	default:
		return "Global"
	}
}

func SkillKindText(kind claudefs.SkillKind) string {
	switch kind {
	case claudefs.SkillKindCommand:
		return "CMD"
	default:
		return "SKILL"
	}
}

func SkillKindStyle(kind claudefs.SkillKind) lipgloss.Style {
	switch kind {
	case claudefs.SkillKindCommand:
		return CommandKindBadgeStyle
	default:
		return SkillKindBadgeStyle
	}
}

func AgentScopeStyle(scope claudefs.AgentScope) lipgloss.Style {
	switch scope {
	case claudefs.AgentScopeProject:
		return ProjectAgentSourceStyle
	case claudefs.AgentScopePlugin:
		return PluginAgentSourceStyle
	default:
		return UserAgentSourceStyle
	}
}

func AgentScopeText(scope claudefs.AgentScope) string {
	switch scope {
	case claudefs.AgentScopeProject:
		return "Project"
	case claudefs.AgentScopePlugin:
		return "Plugin"
	default:
		return "Global"
	}
}
