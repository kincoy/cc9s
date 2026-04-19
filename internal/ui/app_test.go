package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/kincoy/cc9s/internal/claudefs"
)

func TestSwitchToSkillsPreservesProjectContextFromSessions(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceSessions)
	app.sessionList = NewSessionListModelForProject("cc9s")
	app.globalProjectContext = Context{Type: ContextProject, Value: "cc9s"}

	model, cmd := app.Update(SwitchResourceMsg{Resource: ResourceSkills})
	if cmd == nil {
		t.Fatal("expected switch-to-skills command")
	}

	appModel, ok := model.(*AppModel)
	if !ok {
		t.Fatalf("expected *AppModel, got %T", model)
	}

	if appModel.currentResource != ResourceSkills {
		t.Fatalf("resource = %v, want ResourceSkills", appModel.currentResource)
	}
	if appModel.skillList == nil {
		t.Fatal("expected skill list to be initialized")
	}

	ctx := appModel.skillList.GetContext()
	if ctx.Type != ContextProject || ctx.Value != "cc9s" {
		t.Fatalf("skill context = %#v, want project context for cc9s", ctx)
	}
}

func TestSwitchToSkillsFromProjectsDefaultsToAllContext(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceProjects)

	model, cmd := app.Update(SwitchResourceMsg{Resource: ResourceSkills})
	if cmd == nil {
		t.Fatal("expected switch-to-skills command")
	}

	appModel, ok := model.(*AppModel)
	if !ok {
		t.Fatalf("expected *AppModel, got %T", model)
	}

	if appModel.skillList == nil {
		t.Fatal("expected skill list to be initialized")
	}

	ctx := appModel.skillList.GetContext()
	if ctx.Type != ContextAll {
		t.Fatalf("skill context = %#v, want all context", ctx)
	}
}

func TestEscClearsActiveSkillSearchBeforeOtherNavigation(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceSkills)
	app.skillList = NewSkillListModel()
	app.skillList.state.Loading = false
	app.skillList.state.ContextItems = []claudefs.SkillResource{
		{Name: "alpha", Source: claudefs.SkillSourceUser},
		{Name: "beta", Source: claudefs.SkillSourcePlugin},
	}
	app.skillList.ApplyFilter("plugin")

	if len(app.skillList.state.VisibleItems) != 1 {
		t.Fatalf("expected filtered list before esc, got %d items", len(app.skillList.state.VisibleItems))
	}

	model, cmd := app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd != nil {
		t.Fatalf("expected esc clear to avoid emitting command, got %v", cmd)
	}

	appModel, ok := model.(*AppModel)
	if !ok {
		t.Fatalf("expected *AppModel, got %T", model)
	}
	if appModel.skillList.HasActiveFilter() {
		t.Fatal("expected esc to clear active skill filter")
	}
	if len(appModel.skillList.state.VisibleItems) != 2 {
		t.Fatalf("expected full list after esc clear, got %d items", len(appModel.skillList.state.VisibleItems))
	}
}

func TestSwitchToAgentsPreservesProjectContextFromSessions(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceSessions)
	app.sessionList = NewSessionListModelForProject("cc9s")
	app.globalProjectContext = Context{Type: ContextProject, Value: "cc9s"}

	model, cmd := app.Update(SwitchResourceMsg{Resource: ResourceAgents})
	if cmd == nil {
		t.Fatal("expected switch-to-agents command")
	}

	appModel, ok := model.(*AppModel)
	if !ok {
		t.Fatalf("expected *AppModel, got %T", model)
	}
	if appModel.currentResource != ResourceAgents {
		t.Fatalf("resource = %v, want ResourceAgents", appModel.currentResource)
	}
	if appModel.agentList == nil {
		t.Fatal("expected agent list to be initialized")
	}

	ctx := appModel.agentList.GetContext()
	if ctx.Type != ContextProject || ctx.Value != "cc9s" {
		t.Fatalf("agent context = %#v, want project context for cc9s", ctx)
	}
}

func TestSwitchToAgentsFromProjectsDefaultsToAllContext(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceProjects)

	model, cmd := app.Update(SwitchResourceMsg{Resource: ResourceAgents})
	if cmd == nil {
		t.Fatal("expected switch-to-agents command")
	}

	appModel, ok := model.(*AppModel)
	if !ok {
		t.Fatalf("expected *AppModel, got %T", model)
	}
	if appModel.agentList == nil {
		t.Fatal("expected agent list to be initialized")
	}

	ctx := appModel.agentList.GetContext()
	if ctx.Type != ContextAll {
		t.Fatalf("agent context = %#v, want all context", ctx)
	}
}

func TestEscClearsActiveAgentSearchBeforeOtherNavigation(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceAgents)
	app.agentList = NewAgentListModel()
	app.agentList.state.Loading = false
	app.agentList.state.ContextItems = []claudefs.AgentResource{
		{Name: "alpha", Source: claudefs.AgentSourceUser},
		{Name: "beta", Source: claudefs.AgentSourcePlugin},
	}
	app.agentList.ApplyFilter("plugin")

	if len(app.agentList.state.VisibleItems) != 1 {
		t.Fatalf("expected filtered list before esc, got %d items", len(app.agentList.state.VisibleItems))
	}

	model, cmd := app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd != nil {
		t.Fatalf("expected esc clear to avoid emitting command, got %v", cmd)
	}

	appModel, ok := model.(*AppModel)
	if !ok {
		t.Fatalf("expected *AppModel, got %T", model)
	}
	if appModel.agentList.HasActiveFilter() {
		t.Fatal("expected esc to clear active agent filter")
	}
	if len(appModel.agentList.state.VisibleItems) != 2 {
		t.Fatalf("expected full list after esc clear, got %d items", len(appModel.agentList.state.VisibleItems))
	}
}

func TestEscClosesHelpOverlay(t *testing.T) {
	app := NewAppModel()
	app.showHelp = true

	model, cmd := app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd != nil {
		t.Fatalf("expected no command when closing help, got %v", cmd)
	}

	appModel, ok := model.(*AppModel)
	if !ok {
		t.Fatalf("expected *AppModel, got %T", model)
	}
	if appModel.showHelp {
		t.Fatal("expected esc to close help overlay")
	}
}

func TestHelpOverlayScrollsWithJK(t *testing.T) {
	app := NewAppModel()
	app.showHelp = true
	app.helpScroll = 0
	app.height = 20
	app.width = 120

	model, cmd := app.Update(tea.KeyPressMsg{Text: "j", Code: 'j'})
	if cmd != nil {
		t.Fatalf("expected no command when scrolling help, got %v", cmd)
	}

	appModel, ok := model.(*AppModel)
	if !ok {
		t.Fatalf("expected *AppModel, got %T", model)
	}
	if appModel.helpScroll != 1 {
		t.Fatalf("helpScroll = %d, want 1", appModel.helpScroll)
	}

	model, cmd = appModel.Update(tea.KeyPressMsg{Text: "k", Code: 'k'})
	if cmd != nil {
		t.Fatalf("expected no command when scrolling help up, got %v", cmd)
	}

	appModel = model.(*AppModel)
	if appModel.helpScroll != 0 {
		t.Fatalf("helpScroll = %d, want 0", appModel.helpScroll)
	}
}

func TestHelpOverlayScrollDoesNotExceedMax(t *testing.T) {
	app := NewAppModel()
	app.showHelp = true
	app.height = 20
	app.width = 120

	maxScroll := app.maxHelpScroll()
	app.helpScroll = maxScroll

	model, cmd := app.Update(tea.KeyPressMsg{Text: "j", Code: 'j'})
	if cmd != nil {
		t.Fatalf("expected no command when scrolling help at bottom, got %v", cmd)
	}

	appModel := model.(*AppModel)
	if appModel.helpScroll != maxScroll {
		t.Fatalf("helpScroll = %d, want %d", appModel.helpScroll, maxScroll)
	}
}

func TestAgentLoadErrorBlocksSearchMode(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceAgents)
	app.agentList = NewAgentListModel()
	app.agentList.state.Loading = false
	app.agentList.loadErr = assertError("load failed")

	model, cmd := app.Update(tea.KeyPressMsg{Text: "/", Code: '/'})
	if cmd != nil {
		t.Fatalf("expected no command when load error blocks search, got %v", cmd)
	}

	appModel, ok := model.(*AppModel)
	if !ok {
		t.Fatalf("expected *AppModel, got %T", model)
	}
	if appModel.inputMode != InputNormal {
		t.Fatalf("input mode = %v, want InputNormal", appModel.inputMode)
	}
}

func TestShowProjectDetailMessageOpensProjectDetailOverlay(t *testing.T) {
	app := NewAppModel()

	model, cmd := app.Update(ShowProjectDetailMsg{
		Project: claudefs.Project{Name: "cc9s", Path: "/tmp/cc9s"},
	})
	if cmd != nil {
		t.Fatalf("expected no async command for project detail init, got %v", cmd)
	}

	appModel, ok := model.(*AppModel)
	if !ok {
		t.Fatalf("expected *AppModel, got %T", model)
	}
	if !appModel.showingProjectDetail {
		t.Fatal("expected project detail overlay to be visible")
	}
	if appModel.projectDetailView == nil {
		t.Fatal("expected project detail view to be initialized")
	}
	if appModel.projectDetailView.project.Name != "cc9s" {
		t.Fatalf("project detail name = %q, want cc9s", appModel.projectDetailView.project.Name)
	}
}

func TestProjectDetailOverlayClosesViaMessageFlow(t *testing.T) {
	app := NewAppModel()
	app.showingProjectDetail = true
	app.projectDetailView = NewProjectDetailViewModel(claudefs.Project{Name: "cc9s", Path: "/tmp/cc9s"})

	model, cmd := app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected close-project-detail command from overlay")
	}

	appModel := model.(*AppModel)
	if !appModel.showingProjectDetail {
		t.Fatal("expected project detail overlay to remain visible until close message is handled")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(CloseProjectDetailMsg); !ok {
		t.Fatalf("expected CloseProjectDetailMsg, got %T", resultMsg)
	}

	model, nextCmd := appModel.Update(resultMsg)
	if nextCmd != nil {
		t.Fatalf("expected no async command when closing project detail, got %v", nextCmd)
	}

	appModel = model.(*AppModel)
	if appModel.showingProjectDetail {
		t.Fatal("expected project detail overlay to close after close message")
	}
	if appModel.projectDetailView != nil {
		t.Fatal("expected project detail view to be cleared after close message")
	}
}

func TestEnterProjectSyncsActiveResourceToSessions(t *testing.T) {
	app := NewAppModel()
	app.projectList.loading = false
	app.projectList.projects = []claudefs.Project{
		{Name: "cc9s", Path: "/tmp/cc9s"},
	}

	model, cmd := app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected enter-project command")
	}

	appModel, ok := model.(*AppModel)
	if !ok {
		t.Fatalf("expected *AppModel, got %T", model)
	}
	if appModel.currentResource != ResourceProjects {
		t.Fatalf("resource = %v, want ResourceProjects before enter-project message is handled", appModel.currentResource)
	}

	resultMsg := cmd()
	enterMsg, ok := resultMsg.(EnterProjectMsg)
	if !ok {
		t.Fatalf("expected EnterProjectMsg, got %T", resultMsg)
	}

	model, nextCmd := appModel.Update(enterMsg)
	if nextCmd == nil {
		t.Fatal("expected session init command after routing enter-project message")
	}

	appModel = model.(*AppModel)
	if appModel.currentResource != ResourceSessions {
		t.Fatalf("resource = %v, want ResourceSessions", appModel.currentResource)
	}
	if appModel.currentResourceDescriptor().Screen != ScreenSessions {
		t.Fatalf("screen = %v, want ScreenSessions", appModel.currentResourceDescriptor().Screen)
	}
}

func TestBackToProjectsSyncsActiveResourceToProjects(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceSessions)
	app.sessionList = NewSessionListModelForProject("cc9s")
	app.projectList = NewProjectListModel()
	app.lastProjectCursor = 2

	model, cmd := app.Update(BackToProjectsMsg{})
	if cmd != nil {
		t.Fatalf("expected no async command on back to projects, got %v", cmd)
	}

	appModel, ok := model.(*AppModel)
	if !ok {
		t.Fatalf("expected *AppModel, got %T", model)
	}
	if appModel.currentResource != ResourceProjects {
		t.Fatalf("resource = %v, want ResourceProjects", appModel.currentResource)
	}
	if appModel.currentResourceDescriptor().Screen != ScreenProjects {
		t.Fatalf("screen = %v, want ScreenProjects", appModel.currentResourceDescriptor().Screen)
	}
}

func TestSwitchContextMessageUsesDescriptorSetContextForSkills(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceSkills)
	app.skillList = NewSkillListModel()
	app.skillList.state.Loading = false

	model, cmd := app.Update(SwitchContextMsg{Context: Context{Type: ContextProject, Value: "cc9s"}})
	if cmd != nil {
		t.Fatalf("expected no async command, got %v", cmd)
	}

	appModel := model.(*AppModel)
	if got := appModel.skillList.GetContext(); got.Type != ContextProject || got.Value != "cc9s" {
		t.Fatalf("skill context = %#v, want project context for cc9s", got)
	}
}

func TestContextCommandReportsCurrentAgentContextViaDescriptor(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceAgents)
	app.agentList = NewAgentListModel()
	app.agentList.state.Loading = false
	app.agentList.SetContext(Context{Type: ContextProject, Value: "cc9s"})

	cmd := app.executeCommand("context")
	if cmd != nil {
		t.Fatalf("expected no async command when printing current context, got %v", cmd)
	}
	if app.flashMsg != "Current context: cc9s" {
		t.Fatalf("flashMsg = %q, want current agent context", app.flashMsg)
	}
}

func TestCommandCompletionUsesRegistryResourceCommands(t *testing.T) {
	app := NewAppModel()

	candidates, prefix, replaceAll := app.commandCompletionCandidates("ag", false)
	if prefix != "ag" || !replaceAll {
		t.Fatalf("got prefix=%q replaceAll=%v, want ag/true", prefix, replaceAll)
	}
	if len(candidates) != 1 || candidates[0] != "agents" {
		t.Fatalf("candidates = %#v, want [agents]", candidates)
	}
}

func TestCurrentHeaderStateUsesDescriptorForSkillProjectContext(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceSkills)
	app.skillList = NewSkillListModel()
	app.skillList.state.Loading = false
	app.skillList.state.ContextItems = []claudefs.SkillResource{
		{Name: "alpha", Status: claudefs.SkillStatusReady},
	}
	app.skillList.state.VisibleItems = append([]claudefs.SkillResource(nil), app.skillList.state.ContextItems...)
	app.skillList.state.Context = Context{Type: ContextProject, Value: "cc9s"}

	state := app.currentHeaderState()
	if state.ContextLabel != "cc9s" {
		t.Fatalf("context label = %q, want cc9s", state.ContextLabel)
	}
	if state.StatsLabel == "" {
		t.Fatal("expected non-empty stats label")
	}
}

func TestCleanupCommandTogglesSessionCleanupHints(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceSessions)
	app.sessionList = NewSessionListModel()
	app.sessionList.loading = false

	cmd := app.executeCommand("cleanup")
	if cmd == nil {
		t.Fatal("expected cleanup command to emit a toggle message")
	}

	msg := cmd()
	if _, ok := msg.(ToggleCleanupHintsMsg); !ok {
		t.Fatalf("expected ToggleCleanupHintsMsg, got %T", msg)
	}

	model, nextCmd := app.Update(msg)
	if nextCmd != nil {
		t.Fatalf("expected no follow-up command when toggling cleanup hints, got %v", nextCmd)
	}

	appModel := model.(*AppModel)
	if !appModel.sessionList.showCleanupHints {
		t.Fatal("expected cleanup hints to be enabled")
	}
}

func TestCleanupCommandCompletionInSessionsAndProjects(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceSessions)

	candidates, _, _ := app.commandCompletionCandidates("cl", false)
	found := false
	for _, candidate := range candidates {
		if candidate == "cleanup" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected cleanup completion in sessions, got %#v", candidates)
	}

	app.setActiveResource(ResourceProjects)
	candidates, _, _ = app.commandCompletionCandidates("cl", false)
	found = false
	for _, candidate := range candidates {
		if candidate == "cleanup" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected cleanup completion in projects, got %#v", candidates)
	}

	app.setActiveResource(ResourceSkills)
	candidates, _, _ = app.commandCompletionCandidates("cl", false)
	for _, candidate := range candidates {
		if candidate == "cleanup" {
			t.Fatalf("cleanup should not complete outside sessions/projects, got %#v", candidates)
		}
	}
}

func TestCurrentHeaderStateUsesDescriptorForAgentLoadError(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceAgents)
	app.agentList = NewAgentListModel()
	app.agentList.state.Loading = false
	app.agentList.loadErr = assertError("load failed")

	state := app.currentHeaderState()
	if state.StatsLabel != "load error" {
		t.Fatalf("stats label = %q, want load error", state.StatsLabel)
	}
}

// --- Bug 1: StopLoadingMsg must be resource-scoped ---

func TestStopLoadingIgnoredWhenResourceMismatches(t *testing.T) {
	app := NewAppModel()
	app.isLoading = true
	app.loadingResource = ResourceAgents

	// Skills finishes loading — must NOT stop the Agents spinner
	model, cmd := app.Update(StopLoadingMsg{Resource: ResourceSkills})
	if cmd != nil {
		t.Fatalf("expected no command, got %v", cmd)
	}
	appModel := model.(*AppModel)
	if !appModel.isLoading {
		t.Fatal("isLoading should remain true when mismatched resource stops loading")
	}

	// Agents finishes loading — should stop spinner
	model, cmd = appModel.Update(StopLoadingMsg{Resource: ResourceAgents})
	if cmd != nil {
		t.Fatalf("expected no command, got %v", cmd)
	}
	appModel = model.(*AppModel)
	if appModel.isLoading {
		t.Fatal("isLoading should be false when matching resource stops loading")
	}
}

func TestStopLoadingZeroResourceIsUnconditional(t *testing.T) {
	// Health computation failure sends StopLoadingMsg{} (Resource = 0)
	app := NewAppModel()
	app.isLoading = true
	app.loadingResource = ResourceProjects

	model, cmd := app.Update(StopLoadingMsg{})
	if cmd != nil {
		t.Fatalf("expected no command, got %v", cmd)
	}
	appModel := model.(*AppModel)
	if appModel.isLoading {
		t.Fatal("isLoading should be false for unconditional stop (zero Resource)")
	}
}

func TestCrossResourceLoadingRaceDoesNotKillCurrentSpinner(t *testing.T) {
	// Simulate: user on Skills tab, switches to Agents mid-load
	app := NewAppModel()
	app.setActiveResource(ResourceSkills)
	app.skillList = NewSkillListModel()
	app.isLoading = true
	app.loadingResource = ResourceSkills

	// User switches to Agents (starts agents loading)
	model, _ := app.Update(SwitchResourceMsg{Resource: ResourceAgents})
	appModel := model.(*AppModel)
	if !appModel.isLoading {
		t.Fatal("isLoading should be true after switching to Agents")
	}
	if appModel.loadingResource != ResourceAgents {
		t.Fatalf("loadingResource = %v, want ResourceAgents", appModel.loadingResource)
	}

	// Skills finishes loading — must not stop Agents spinner
	model, _ = appModel.Update(skillsLoadedMsg{
		result: claudefs.SkillScanResult{},
	})
	appModel = model.(*AppModel)
	if !appModel.isLoading {
		t.Fatal("isLoading should remain true: Agents is still loading")
	}
}

// --- Bug 2: Forward path must sync detail overlays ---

func TestSkillLoadedViaForwardPathSyncsDetailOverlay(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceSkills)
	app.skillList = NewSkillListModel()
	app.skillList.Init() // Initialize viewport
	app.skillList.state.Loading = true

	// Open detail overlay with old data
	app.showingSkillDetail = true
	app.skillDetailView = NewSkillDetailViewModel(claudefs.SkillResource{
		Name:    "my-skill",
		Path:    "/old/path.md",
		Summary: "old summary",
	})

	// Reload completes with updated data
	model, cmd := app.Update(skillsLoadedMsg{
		result: claudefs.SkillScanResult{
			Skills: []claudefs.SkillResource{
				{Name: "my-skill", Path: "/new/path.md", Summary: "new summary"},
			},
		},
	})
	if cmd == nil {
		t.Fatal("expected StopLoadingMsg command from forward path")
	}

	appModel := model.(*AppModel)
	if appModel.skillDetailView.skill.Path != "/new/path.md" {
		t.Fatalf("detail overlay skill path = %q, want /new/path.md (overlay should sync after reload)",
			appModel.skillDetailView.skill.Path)
	}
	if appModel.skillDetailView.skill.Summary != "new summary" {
		t.Fatalf("detail overlay skill summary = %q, want 'new summary'",
			appModel.skillDetailView.skill.Summary)
	}
}

func TestAgentLoadedViaForwardPathSyncsDetailOverlay(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceAgents)
	app.agentList = NewAgentListModel()
	app.agentList.Init() // Initialize viewport
	app.agentList.state.Loading = true

	// Open detail overlay with old data
	app.showingAgentDetail = true
	app.agentDetailView = NewAgentDetailViewModel(claudefs.AgentResource{
		Name:    "my-agent",
		Path:    "/old/path.md",
		Summary: "old summary",
	})

	// Reload completes with updated data
	model, cmd := app.Update(agentsLoadedMsg{
		result: claudefs.AgentScanResult{
			Agents: []claudefs.AgentResource{
				{Name: "my-agent", Path: "/new/path.md", Summary: "new summary"},
			},
		},
	})
	if cmd == nil {
		t.Fatal("expected StopLoadingMsg command from forward path")
	}

	appModel := model.(*AppModel)
	if appModel.agentDetailView.agent.Path != "/new/path.md" {
		t.Fatalf("detail overlay agent path = %q, want /new/path.md (overlay should sync after reload)",
			appModel.agentDetailView.agent.Path)
	}
	if appModel.agentDetailView.agent.Summary != "new summary" {
		t.Fatalf("detail overlay agent summary = %q, want 'new summary'",
			appModel.agentDetailView.agent.Summary)
	}
}

func TestSkillListUpdateReturnsStopLoadingMsgCommand(t *testing.T) {
	// Simple unit test to verify SkillListModel.Update returns a cmd
	skillList := NewSkillListModel()
	skillList.Init()
	skillList.state.Loading = true

	cmd := skillList.Update(skillsLoadedMsg{
		result: claudefs.SkillScanResult{
			Skills: []claudefs.SkillResource{
				{Name: "test-skill", Path: "/test/path.md"},
			},
		},
	})

	if cmd == nil {
		t.Fatal("skillList.Update should return StopLoadingMsg command")
	}
}

func TestAppUpdateForwardsSkillLoadedAndReturnsCmd(t *testing.T) {
	// Test that app.Update correctly forwards skillsLoadedMsg and returns the cmd
	app := NewAppModel()
	app.setActiveResource(ResourceSkills)
	app.skillList = NewSkillListModel()
	app.skillList.Init()
	app.skillList.state.Loading = true

	// Add detail overlay to match original test scenario
	app.showingSkillDetail = true
	app.skillDetailView = NewSkillDetailViewModel(claudefs.SkillResource{
		Name:    "test-skill",
		Path:    "/old/path.md",
		Summary: "old",
	})

	model, cmd := app.Update(skillsLoadedMsg{
		result: claudefs.SkillScanResult{
			Skills: []claudefs.SkillResource{
				{Name: "test-skill", Path: "/new/path.md", Summary: "new"},
			},
		},
	})

	if cmd == nil {
		t.Fatal("app.Update should forward skillsLoadedMsg and return StopLoadingMsg command")
	}

	appModel := model.(*AppModel)
	if appModel.skillList.state.Loading {
		t.Fatal("skillList should have Loading=false after processing loaded message")
	}
}

type assertError string

func (e assertError) Error() string { return string(e) }
