package ui

import tea "charm.land/bubbletea/v2"

// ResourceCapabilities captures the shared interactions a resource supports.
type ResourceCapabilities struct {
	SupportsSearch             bool
	SupportsContext            bool
	SupportsDetail             bool
	SupportsEdit               bool
	SupportsLog                bool
	SupportsAllContextShortcut bool
}

// ResourceHeaderState is the normalized header payload for a resource page.
type ResourceHeaderState struct {
	ContextLabel     string
	StatsLabel       string
	FilteredCount    int
	TotalCount       int
	HasFilteredState bool
}

// ResourceDescriptor defines how a top-level resource participates in the UI model.
type ResourceDescriptor struct {
	Resource             ResourceType
	Screen               Screen
	CommandName          string
	DisplayName          string
	Capabilities         ResourceCapabilities
	FooterHints          func(FooterContext) []KeyHint
	HelpSection          func() ResourceHelpSection
	ResolveTargetContext func(*AppModel) Context
	EnsureActive         func(*AppModel, Context) tea.Cmd
	CurrentContext       func(*AppModel) Context
	SetContext           func(*AppModel, Context) tea.Cmd
	ApplyFilter          func(*AppModel, string)
	HasActiveFilter      func(*AppModel) bool
	CanStartSearch       func(*AppModel) bool
	HeaderState          func(*AppModel) ResourceHeaderState
}
