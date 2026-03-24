package ui

import (
	"testing"

	"github.com/kincoy/cc9s/internal/claudefs"
)

func TestAgentListSearchNormalizesSlashPrefix(t *testing.T) {
	model := &AgentListModel{
		contextAgents: []claudefs.AgentResource{
			{Name: "alpha", Source: claudefs.AgentSourceUser, Scope: claudefs.AgentScopeUser},
			{Name: "beta", Source: claudefs.AgentSourceProject, Scope: claudefs.AgentScopeProject},
		},
	}

	model.ApplyFilter("/user")
	if len(model.agents) != 1 {
		t.Fatalf("expected 1 filtered agent, got %d", len(model.agents))
	}
	if model.agents[0].Source != claudefs.AgentSourceUser {
		t.Fatalf("expected user agent, got %s", model.agents[0].Source)
	}
}

func TestAgentListSearchMatchesVisibleAliasesAndReasons(t *testing.T) {
	model := &AgentListModel{
		contextAgents: []claudefs.AgentResource{
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
	}

	model.ApplyFilter("/global")
	if len(model.agents) != 1 || model.agents[0].Source != claudefs.AgentSourceUser {
		t.Fatalf("global search mismatch: %+v", model.agents)
	}

	model.ApplyFilter("/recognized by claude")
	if len(model.agents) != 1 || model.agents[0].Source != claudefs.AgentSourcePlugin {
		t.Fatalf("reason search mismatch: %+v", model.agents)
	}
}

func TestAgentListSortRebuildsCurrentView(t *testing.T) {
	model := &AgentListModel{
		context: Context{Type: ContextAll},
		allAgents: []claudefs.AgentResource{
			{Name: "beta", Path: "/tmp/beta.md", Status: claudefs.AgentStatusInvalid},
			{Name: "alpha", Path: "/tmp/alpha.md", Status: claudefs.AgentStatusReady},
		},
		sortBy:  SortByAgentName,
		sortAsc: true,
	}

	model.applyContext()
	model.sortBy = SortByAgentStatus
	model.sortAsc = true
	model.sortAgents(model.allAgents)
	model.applyContext()

	if len(model.agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(model.agents))
	}
	if model.agents[0].Status != claudefs.AgentStatusInvalid {
		t.Fatalf("expected invalid agent first after status sort, got %s", model.agents[0].Status)
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
