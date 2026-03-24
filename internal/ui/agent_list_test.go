package ui

import (
	"testing"

	"github.com/kincoy/cc9s/internal/claudefs"
)

func TestAgentListSearchNormalizesSlashPrefix(t *testing.T) {
	model := &AgentListModel{
		state: DefaultResourceListState[claudefs.AgentResource]{
			ContextItems: []claudefs.AgentResource{
				{Name: "alpha", Source: claudefs.AgentSourceUser, Scope: claudefs.AgentScopeUser},
				{Name: "beta", Source: claudefs.AgentSourceProject, Scope: claudefs.AgentScopeProject},
			},
		},
	}

	model.ApplyFilter("/user")
	if len(model.state.VisibleItems) != 1 {
		t.Fatalf("expected 1 filtered agent, got %d", len(model.state.VisibleItems))
	}
	if model.state.VisibleItems[0].Source != claudefs.AgentSourceUser {
		t.Fatalf("expected user agent, got %s", model.state.VisibleItems[0].Source)
	}
}

func TestAgentListSearchMatchesVisibleAliasesAndReasons(t *testing.T) {
	model := &AgentListModel{
		state: DefaultResourceListState[claudefs.AgentResource]{
			ContextItems: []claudefs.AgentResource{
				{
					Name:              "go-reviewer",
					Source:            claudefs.AgentSourcePlugin,
					Scope:             claudefs.AgentScopePlugin,
					PluginName:        "everything-claude-code",
					ValidationReasons: []string{"Agent file is not recognized by Claude Code"},
				},
				{
					Name:   "helper",
					Source: claudefs.AgentSourceUser,
					Scope:  claudefs.AgentScopeUser,
				},
			},
		},
	}

	model.ApplyFilter("/global")
	if len(model.state.VisibleItems) != 1 || model.state.VisibleItems[0].Source != claudefs.AgentSourceUser {
		t.Fatalf("global search mismatch: %+v", model.state.VisibleItems)
	}

	model.ApplyFilter("/recognized by claude")
	if len(model.state.VisibleItems) != 1 || model.state.VisibleItems[0].Source != claudefs.AgentSourcePlugin {
		t.Fatalf("reason search mismatch: %+v", model.state.VisibleItems)
	}
}

func TestAgentListSortRebuildsCurrentView(t *testing.T) {
	model := &AgentListModel{
		state: DefaultResourceListState[claudefs.AgentResource]{
			Context: Context{Type: ContextAll},
			AllItems: []claudefs.AgentResource{
				{Name: "beta", Path: "/tmp/beta.md", Status: claudefs.AgentStatusInvalid},
				{Name: "alpha", Path: "/tmp/alpha.md", Status: claudefs.AgentStatusReady},
			},
		},
		sortBy:  SortByAgentName,
		sortAsc: true,
	}

	model.applyContext()
	model.sortBy = SortByAgentStatus
	model.sortAsc = true
	model.sortAgents(model.state.AllItems)
	model.applyContext()

	if len(model.state.VisibleItems) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(model.state.VisibleItems))
	}
	if model.state.VisibleItems[0].Status != claudefs.AgentStatusInvalid {
		t.Fatalf("expected invalid agent first after status sort, got %s", model.state.VisibleItems[0].Status)
	}
}

func TestAgentSortCycleFollowsVisibleColumnOrder(t *testing.T) {
	model := &AgentListModel{sortBy: SortByAgentName}

	model.sortBy = (model.sortBy + 1) % 3
	if model.sortBy != SortByAgentStatus {
		t.Fatalf("first cycle = %v, want SortByAgentStatus", model.sortBy)
	}

	model.sortBy = (model.sortBy + 1) % 3
	if model.sortBy != SortByAgentScope {
		t.Fatalf("second cycle = %v, want SortByAgentScope", model.sortBy)
	}
}

func TestAgentListReloadPreservesFocusedAgent(t *testing.T) {
	model := &AgentListModel{
		state: DefaultResourceListState[claudefs.AgentResource]{
			Context: Context{Type: ContextAll},
			VisibleItems: []claudefs.AgentResource{
				{Name: "alpha", Path: "/tmp/alpha.md", Source: claudefs.AgentSourceUser},
				{Name: "beta", Path: "/tmp/beta.md", Source: claudefs.AgentSourceProject},
				{Name: "gamma", Path: "/tmp/gamma.md", Source: claudefs.AgentSourcePlugin},
			},
			Cursor: 1,
		},
	}

	model.captureCursorForReload()
	model.Update(agentsLoadedMsg{
		result: claudefs.AgentScanResult{
			Agents: []claudefs.AgentResource{
				{Name: "gamma", Path: "/tmp/gamma.md", Source: claudefs.AgentSourcePlugin},
				{Name: "beta", Path: "/tmp/beta.md", Source: claudefs.AgentSourceProject},
				{Name: "alpha", Path: "/tmp/alpha.md", Source: claudefs.AgentSourceUser},
			},
		},
	})

	if model.state.Cursor != 1 {
		t.Fatalf("cursor = %d, want 1", model.state.Cursor)
	}
	if model.state.VisibleItems[model.state.Cursor].Name != "beta" {
		t.Fatalf("focused agent = %s, want beta", model.state.VisibleItems[model.state.Cursor].Name)
	}
}

func TestAgentSetContextClearsActiveFilter(t *testing.T) {
	model := &AgentListModel{
		state: DefaultResourceListState[claudefs.AgentResource]{
			Context: Context{Type: ContextAll},
			ContextItems: []claudefs.AgentResource{
				{Name: "alpha", Source: claudefs.AgentSourceUser},
			},
			FilterQuery: "alpha",
		},
	}

	model.SetContext(Context{Type: ContextProject, Value: "cc9s"})

	if model.state.FilterQuery != "" {
		t.Fatalf("filterQuery = %q, want empty", model.state.FilterQuery)
	}
}
