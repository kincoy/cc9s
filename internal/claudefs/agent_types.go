package claudefs

import "time"

// AgentSource identifies where an agent was discovered.
type AgentSource string

const (
	AgentSourceUser    AgentSource = "user"
	AgentSourceProject AgentSource = "project"
	AgentSourcePlugin  AgentSource = "plugin"
)

// AgentScope is the user-facing scope label for an agent.
type AgentScope string

const (
	AgentScopeUser    AgentScope = "User"
	AgentScopeProject AgentScope = "Project"
	AgentScopePlugin  AgentScope = "Plugin"
)

// AgentStatus is the lightweight readiness state exposed to users.
type AgentStatus string

const (
	AgentStatusReady   AgentStatus = "Ready"
	AgentStatusInvalid AgentStatus = "Invalid"
)

// AgentAvailabilityResult captures the user-facing readiness outcome and reasons.
type AgentAvailabilityResult struct {
	Status  AgentStatus
	Reasons []string
}

// AgentDiscoveryRoot represents a configured scan root.
type AgentDiscoveryRoot struct {
	Path              string
	Source            AgentSource
	ProjectName       string
	ProjectPath       string
	PluginName        string
	PluginInstallMode string
}

// AgentResource is a discovered local agent surfaced in the TUI.
type AgentResource struct {
	Name              string
	Path              string
	Source            AgentSource
	Scope             AgentScope
	ProjectName       string
	ProjectPath       string
	PluginName        string
	PluginInstallMode string
	Status            AgentStatus
	Summary           string
	Description       string
	Model             string
	Tools             []string
	PermissionMode    string
	Memory            string
	ValidationReasons []string
	Availability      AgentAvailabilityResult
}

// AgentRecognitionSnapshot holds the Claude-recognized agent names for one scope.
type AgentRecognitionSnapshot struct {
	CustomAgents map[string]struct{}
	PluginAgents map[string]struct{}
}

// AgentScanResult holds the results of an agent scan.
type AgentScanResult struct {
	Agents       []AgentResource
	ReadyCount   int
	InvalidCount int
	ScanDuration time.Duration
	Err          error
}
