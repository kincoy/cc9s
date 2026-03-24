package claudefs

import "time"

// SkillSource identifies where a skill was discovered.
type SkillSource string

const (
	SkillSourceUser    SkillSource = "user"
	SkillSourceProject SkillSource = "project"
	SkillSourcePlugin  SkillSource = "plugin"
)

// SkillScope is the user-facing scope label for a skill.
type SkillScope string

const (
	SkillScopeUser    SkillScope = "User"
	SkillScopeProject SkillScope = "Project"
	SkillScopePlugin  SkillScope = "Plugin"
)

// SkillKind distinguishes skills from commands surfaced in the same resource view.
type SkillKind string

const (
	SkillKindSkill   SkillKind = "Skill"
	SkillKindCommand SkillKind = "Command"
)

// SkillStatus is the lightweight readiness state exposed to users.
type SkillStatus string

const (
	SkillStatusReady   SkillStatus = "Ready"
	SkillStatusInvalid SkillStatus = "Invalid"
)

// SkillShape describes how the skill is represented on disk.
type SkillShape string

const (
	SkillShapeDirectory  SkillShape = "DirectorySkill"
	SkillShapeSingleFile SkillShape = "SingleFileSkill"
)

// SkillAvailabilityResult captures the user-facing readiness outcome and reasons.
type SkillAvailabilityResult struct {
	Status  SkillStatus
	Reasons []string
}

// SkillDiscoveryRoot represents a configured scan root.
type SkillDiscoveryRoot struct {
	Path              string
	Source            SkillSource
	Kind              SkillKind
	ProjectName       string
	ProjectPath       string
	PluginName        string
	PluginInstallMode string
}

// SkillResource is a discovered local skill surfaced in the TUI.
type SkillResource struct {
	Name              string
	Path              string
	Source            SkillSource
	Scope             SkillScope
	Kind              SkillKind
	ProjectName       string
	ProjectPath       string
	PluginName        string
	PluginInstallMode string
	Status            SkillStatus
	Summary           string
	EntryFile         string
	AssociatedFiles   []string
	ValidationReasons []string
	Shape             SkillShape
}

// SkillScanResult holds the results of a skill scan.
type SkillScanResult struct {
	Skills       []SkillResource
	ReadyCount   int
	InvalidCount int
	ScanDuration time.Duration
	Err          error
}
