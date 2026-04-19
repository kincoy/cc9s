package ui

import tea "charm.land/bubbletea/v2"

// ResourceRegistry is the source of truth for top-level resource descriptors.
type ResourceRegistry struct {
	ordered   []ResourceDescriptor
	byType    map[ResourceType]ResourceDescriptor
	byCommand map[string]ResourceDescriptor
}

func newResourceRegistry() *ResourceRegistry {
	ordered := []ResourceDescriptor{
		{
			Resource:    ResourceProjects,
			Screen:      ScreenProjects,
			CommandName: "projects",
			DisplayName: "Projects",
			Capabilities: ResourceCapabilities{
				SupportsSearch: true,
				SupportsDetail: true,
			},
			FooterHints: func(ctx FooterContext) []KeyHint {
				hints := []KeyHint{
					{Key: "q", Label: "Quit"},
					{Key: "j/k", Label: "Navigate"},
					{Key: "s/S", Label: "Sort"},
					{Key: "r", Label: "Refresh"},
					{Key: "Enter", Label: "Open"},
					{Key: ":", Label: "Cmd"},
					{Key: "?", Label: "Help"},
				}
				if ctx.Descriptor.Capabilities.SupportsDetail {
					hints = append(hints, KeyHint{Key: "d", Label: "Detail"})
				}
				if ctx.Descriptor.Capabilities.SupportsSearch {
					hints = append(hints, KeyHint{Key: "/", Label: "Search"})
				}
				return hints
			},
			HelpSection: func() ResourceHelpSection {
				return ResourceHelpSection{
					Title: "Project Operations",
					Lines: []KeyHint{
						{Key: ":projects", Label: "    Switch to projects resource"},
						{Key: ":health", Label: "      Toggle HEALTH column"},
						{Key: ":cleanup", Label: "     Toggle STATUS column (deleted projects)"},
						{Key: "Enter", Label: "     Open selected project sessions"},
						{Key: "d", Label: "         View selected project details"},
						{Key: "r", Label: "         Refresh data from disk"},
						{Key: "/", Label: "         Search projects by name or path"},
					},
				}
			},
			ResolveTargetContext: func(_ *AppModel) Context {
				return Context{Type: ContextAll}
			},
			EnsureActive: func(a *AppModel, _ Context) tea.Cmd {
				a.setActiveResource(ResourceProjects)
				a.globalProjectContext = Context{Type: ContextAll}
				return nil
			},
			CurrentContext: func(_ *AppModel) Context {
				return Context{Type: ContextAll}
			},
			ApplyFilter: func(a *AppModel, query string) {
				if a.projectList != nil {
					a.projectList.ApplyFilter(query)
				}
			},
			HasActiveFilter: func(a *AppModel) bool {
				return a.projectList != nil && a.projectList.HasActiveFilter()
			},
			CanStartSearch: func(a *AppModel) bool {
				return a.projectList != nil
			},
			HeaderState: func(a *AppModel) ResourceHeaderState {
				projectCount, totalSessions, activeCount := a.projectList.GetStats()
				state := ResourceHeaderState{
					ContextLabel: "",
					StatsLabel:   formatProjectSummary(projectCount, totalSessions, activeCount),
				}
				if a.inputMode == InputSearch {
					state.FilteredCount, state.TotalCount = a.projectList.GetFilterStats()
					state.HasFilteredState = true
				}
				return state
			},
		},
		{
			Resource:    ResourceSessions,
			Screen:      ScreenSessions,
			CommandName: "sessions",
			DisplayName: "Sessions",
			Capabilities: ResourceCapabilities{
				SupportsSearch:             true,
				SupportsContext:            true,
				SupportsDetail:             true,
				SupportsLog:                true,
				SupportsAllContextShortcut: true,
			},
			FooterHints: func(ctx FooterContext) []KeyHint {
				hints := []KeyHint{
					{Key: "q", Label: "Quit"},
					{Key: "j/k", Label: "Navigate"},
					{Key: "s/S", Label: "Sort"},
					{Key: "r", Label: "Refresh"},
					{Key: "Enter", Label: "Resume"},
					{Key: "Space", Label: "Select"},
					{Key: ":", Label: "Cmd"},
					{Key: "Esc", Label: "Back"},
				}
				if ctx.Descriptor.Capabilities.SupportsDetail {
					hints = append(hints, KeyHint{Key: "d", Label: "Detail"})
				}
				if ctx.Descriptor.Capabilities.SupportsLog {
					hints = append(hints, KeyHint{Key: "l", Label: "Logs"})
				}
				if ctx.Descriptor.Capabilities.SupportsSearch {
					hints = append(hints, KeyHint{Key: "/", Label: "Search"})
				}
				if ctx.HasMulti {
					hints = append(hints, KeyHint{Key: "x", Label: "Delete"})
				}
				if ctx.Descriptor.Capabilities.SupportsAllContextShortcut {
					hints = append(hints, KeyHint{Key: "0", Label: "All ctx"})
				}
				hints = append(hints, KeyHint{Key: "?", Label: "Help"})
				return hints
			},
			HelpSection: func() ResourceHelpSection {
				return ResourceHelpSection{
					Title: "Session Operations",
					Lines: []KeyHint{
						{Key: ":sessions", Label: "    Switch to sessions resource"},
						{Key: ":cleanup", Label: "   Toggle cleanup recommendations"},
						{Key: "Enter", Label: "     Resume selected session"},
						{Key: "d", Label: "         View session details"},
						{Key: "l", Label: "         View session log"},
						{Key: "r", Label: "         Refresh data from disk"},
						{Key: "/", Label: "         Search sessions or lifecycle states"},
						{Key: "Esc", Label: "       Go back to projects"},
						{Key: "Space", Label: "      Toggle select session"},
						{Key: "x", Label: "         Delete selected session(s)"},
					},
				}
			},
			ResolveTargetContext: func(a *AppModel) Context {
				return a.globalProjectContext
			},
			EnsureActive: func(a *AppModel, targetCtx Context) tea.Cmd {
				a.setActiveResource(ResourceSessions)
				if a.sessionList == nil {
					a.sessionList = NewSessionListModel()
					cmd := a.sessionList.Init()
					if a.ready {
						a.sessionList.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
					}
					// Set loading state synchronously to avoid race with async scan
					a.isLoading = true
					a.loadingText = "Loading sessions..."
					a.loadingResource = ResourceSessions
					if ctxCmd := a.sessionList.SetContext(targetCtx); ctxCmd != nil {
						return tea.Batch(cmd, ctxCmd, a.spinner.Tick)
					}
					return tea.Batch(cmd, a.spinner.Tick)
				}
				return a.sessionList.SetContext(targetCtx)
			},
			CurrentContext: func(a *AppModel) Context {
				return a.contextForResource(ResourceSessions)
			},
			SetContext: func(a *AppModel, ctx Context) tea.Cmd {
				if a.sessionList == nil {
					return nil
				}
				a.setActiveResource(ResourceSessions)
				return a.sessionList.SetContext(ctx)
			},
			ApplyFilter: func(a *AppModel, query string) {
				if a.sessionList != nil {
					a.sessionList.ApplyFilter(query)
				}
			},
			HasActiveFilter: func(a *AppModel) bool {
				return a.sessionList != nil && a.sessionList.HasActiveFilter()
			},
			CanStartSearch: func(a *AppModel) bool {
				return a.sessionList != nil
			},
			HeaderState: func(a *AppModel) ResourceHeaderState {
				if a.sessionList == nil {
					return ResourceHeaderState{}
				}
				ctx := a.sessionList.GetContext()
				summary := summarizeGlobalSessions(a.sessionList.sessions)
				state := ResourceHeaderState{
					ContextLabel: formatResourceContextLabel("All Projects", ctx),
					StatsLabel:   formatLifecycleSummary(a.width, summary),
				}
				if a.inputMode == InputSearch {
					state.FilteredCount, state.TotalCount = a.sessionList.GetFilterStats()
					state.HasFilteredState = true
				}
				return state
			},
		},
		{
			Resource:    ResourceSkills,
			Screen:      ScreenSkills,
			CommandName: "skills",
			DisplayName: "Skills",
			Capabilities: ResourceCapabilities{
				SupportsSearch:             true,
				SupportsContext:            true,
				SupportsDetail:             true,
				SupportsEdit:               true,
				SupportsAllContextShortcut: true,
			},
			FooterHints: func(ctx FooterContext) []KeyHint {
				return defaultResourceFooterHints(ctx)
			},
			HelpSection: func() ResourceHelpSection {
				return ResourceHelpSection{
					Title: "Skill Operations",
					Lines: []KeyHint{
						{Key: ":skills", Label: "    Switch to skills resource"},
						{Key: "d", Label: "         View selected skill or command details"},
						{Key: "e", Label: "         Edit selected skill or command"},
						{Key: "r", Label: "         Refresh data from disk"},
						{Key: "/", Label: "         Search skills and commands by name, path, scope, status"},
					},
				}
			},
			ResolveTargetContext: func(a *AppModel) Context {
				return a.globalProjectContext
			},
			EnsureActive: func(a *AppModel, targetCtx Context) tea.Cmd {
				a.setActiveResource(ResourceSkills)
				if a.skillList == nil {
					a.skillList = NewSkillListModel()
					cmd := a.skillList.Init()
					if a.ready {
						a.skillList.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
					}
					a.isLoading = true
					a.loadingText = "Loading skills..."
					a.loadingResource = ResourceSkills
					if ctxCmd := a.skillList.SetContext(targetCtx); ctxCmd != nil {
						return tea.Batch(cmd, ctxCmd, a.spinner.Tick)
					}
					return tea.Batch(cmd, a.spinner.Tick)
				}
				return a.skillList.SetContext(targetCtx)
			},
			CurrentContext: func(a *AppModel) Context {
				return a.contextForResource(ResourceSkills)
			},
			SetContext: func(a *AppModel, ctx Context) tea.Cmd {
				if a.skillList == nil {
					return nil
				}
				a.setActiveResource(ResourceSkills)
				return a.skillList.SetContext(ctx)
			},
			ApplyFilter: func(a *AppModel, query string) {
				if a.skillList != nil {
					a.skillList.ApplyFilter(query)
				}
			},
			HasActiveFilter: func(a *AppModel) bool {
				return a.skillList != nil && a.skillList.HasActiveFilter()
			},
			CanStartSearch: func(a *AppModel) bool {
				return a.skillList != nil
			},
			HeaderState: func(a *AppModel) ResourceHeaderState {
				if a.skillList == nil {
					return ResourceHeaderState{}
				}
				total, ready, invalid := a.skillList.GetStats()
				ctx := a.skillList.GetContext()
				state := ResourceHeaderState{
					ContextLabel: formatResourceContextLabel("All Skills", ctx),
					StatsLabel:   formatSkillSummary(a.width, total, ready, invalid),
				}
				if a.inputMode == InputSearch {
					state.FilteredCount, state.TotalCount = a.skillList.GetFilterStats()
					state.HasFilteredState = true
				}
				return state
			},
		},
		{
			Resource:    ResourceAgents,
			Screen:      ScreenAgents,
			CommandName: "agents",
			DisplayName: "Agents",
			Capabilities: ResourceCapabilities{
				SupportsSearch:             true,
				SupportsContext:            true,
				SupportsDetail:             true,
				SupportsEdit:               true,
				SupportsAllContextShortcut: true,
			},
			FooterHints: func(ctx FooterContext) []KeyHint {
				return defaultResourceFooterHints(ctx)
			},
			HelpSection: func() ResourceHelpSection {
				return ResourceHelpSection{
					Title: "Agent Operations",
					Lines: []KeyHint{
						{Key: ":agents", Label: "    Switch to agents resource"},
						{Key: "d", Label: "         View selected agent details"},
						{Key: "e", Label: "         Edit selected agent file"},
						{Key: "r", Label: "         Refresh data from disk"},
						{Key: "/", Label: "         Search agents by name, path, scope, status, config"},
					},
				}
			},
			ResolveTargetContext: func(a *AppModel) Context {
				return a.globalProjectContext
			},
			EnsureActive: func(a *AppModel, targetCtx Context) tea.Cmd {
				a.setActiveResource(ResourceAgents)
				if a.agentList == nil {
					a.agentList = NewAgentListModel()
					cmd := a.agentList.Init()
					if a.ready {
						a.agentList.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
					}
					// Set loading state synchronously to avoid race with async scan
					a.isLoading = true
					a.loadingText = "Loading agents..."
					a.loadingResource = ResourceAgents
					if ctxCmd := a.agentList.SetContext(targetCtx); ctxCmd != nil {
						return tea.Batch(cmd, ctxCmd, a.spinner.Tick)
					}
					return tea.Batch(cmd, a.spinner.Tick)
				}
				return a.agentList.SetContext(targetCtx)
			},
			CurrentContext: func(a *AppModel) Context {
				return a.contextForResource(ResourceAgents)
			},
			SetContext: func(a *AppModel, ctx Context) tea.Cmd {
				if a.agentList == nil {
					return nil
				}
				a.setActiveResource(ResourceAgents)
				return a.agentList.SetContext(ctx)
			},
			ApplyFilter: func(a *AppModel, query string) {
				if a.agentList != nil {
					a.agentList.ApplyFilter(query)
				}
			},
			HasActiveFilter: func(a *AppModel) bool {
				return a.agentList != nil && a.agentList.HasActiveFilter()
			},
			CanStartSearch: func(a *AppModel) bool {
				return a.agentList != nil && !a.agentList.HasLoadError()
			},
			HeaderState: func(a *AppModel) ResourceHeaderState {
				if a.agentList == nil {
					return ResourceHeaderState{}
				}
				total, ready, invalid := a.agentList.GetStats()
				ctx := a.agentList.GetContext()
				state := ResourceHeaderState{
					ContextLabel: formatResourceContextLabel("All Agents", ctx),
					StatsLabel:   formatAgentSummary(a.width, total, ready, invalid),
				}
				if a.agentList.HasLoadError() {
					state.StatsLabel = "load error"
				}
				if a.inputMode == InputSearch {
					state.FilteredCount, state.TotalCount = a.agentList.GetFilterStats()
					state.HasFilteredState = true
				}
				return state
			},
		},
	}

	registry := &ResourceRegistry{
		ordered:   ordered,
		byType:    make(map[ResourceType]ResourceDescriptor, len(ordered)),
		byCommand: make(map[string]ResourceDescriptor, len(ordered)),
	}
	for _, descriptor := range ordered {
		registry.byType[descriptor.Resource] = descriptor
		registry.byCommand[descriptor.CommandName] = descriptor
	}
	return registry
}

func (r *ResourceRegistry) MustGet(resource ResourceType) ResourceDescriptor {
	descriptor, ok := r.byType[resource]
	if !ok {
		panic("missing resource descriptor")
	}
	return descriptor
}

func (r *ResourceRegistry) FindByCommand(command string) (ResourceDescriptor, bool) {
	descriptor, ok := r.byCommand[command]
	return descriptor, ok
}

func (r *ResourceRegistry) CommandNames() []string {
	names := make([]string, 0, len(r.ordered))
	for _, descriptor := range r.ordered {
		names = append(names, descriptor.CommandName)
	}
	return names
}

func (r *ResourceRegistry) CompletionCandidates(prefix string) []string {
	names := r.CommandNames()
	candidates := names[:0]
	for _, name := range names {
		if len(prefix) == 0 || (len(name) >= len(prefix) && name[:len(prefix)] == prefix) {
			candidates = append(candidates, name)
		}
	}
	return candidates
}

func defaultResourceFooterHints(ctx FooterContext) []KeyHint {
	hints := []KeyHint{
		{Key: "q", Label: "Quit"},
		{Key: "j/k", Label: "Navigate"},
		{Key: "s/S", Label: "Sort"},
		{Key: "r", Label: "Refresh"},
		{Key: ":", Label: "Cmd"},
		{Key: "?", Label: "Help"},
	}
	if ctx.Descriptor.Capabilities.SupportsDetail {
		hints = append(hints, KeyHint{Key: "d", Label: "Detail"})
	}
	if ctx.Descriptor.Capabilities.SupportsEdit {
		hints = append(hints, KeyHint{Key: "e", Label: "Edit"})
	}
	if ctx.Descriptor.Capabilities.SupportsSearch {
		hints = append(hints, KeyHint{Key: "/", Label: "Search"})
	}
	if ctx.Descriptor.Capabilities.SupportsAllContextShortcut {
		hints = append(hints, KeyHint{Key: "0", Label: "All ctx"})
	}
	return hints
}
