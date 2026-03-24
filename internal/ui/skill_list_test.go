package ui

import (
	"testing"

	"github.com/kincoy/cc9s/internal/claudefs"
)

func TestSkillListSearchNormalizesSlashPrefix(t *testing.T) {
	model := &SkillListModel{
		state: DefaultResourceListState[claudefs.SkillResource]{
			ContextItems: []claudefs.SkillResource{
				{Name: "alpha", Source: claudefs.SkillSourceUser, Scope: claudefs.SkillScopeUser},
				{Name: "beta", Source: claudefs.SkillSourceProject, Scope: claudefs.SkillScopeProject},
			},
		},
	}

	model.ApplyFilter("/user")

	if len(model.state.VisibleItems) != 1 {
		t.Fatalf("expected 1 filtered skill, got %d", len(model.state.VisibleItems))
	}
	if model.state.VisibleItems[0].Source != claudefs.SkillSourceUser {
		t.Fatalf("expected user skill, got %s", model.state.VisibleItems[0].Source)
	}
}

func TestSkillListSearchMatchesVisibleAliases(t *testing.T) {
	model := &SkillListModel{
		state: DefaultResourceListState[claudefs.SkillResource]{
			ContextItems: []claudefs.SkillResource{
				{
					Name:       "security-scan",
					Source:     claudefs.SkillSourcePlugin,
					Scope:      claudefs.SkillScopePlugin,
					Kind:       claudefs.SkillKindCommand,
					PluginName: "everything-claude-code",
				},
				{
					Name:   "session-save",
					Source: claudefs.SkillSourceUser,
					Scope:  claudefs.SkillScopeUser,
					Kind:   claudefs.SkillKindSkill,
				},
			},
		},
	}

	model.ApplyFilter("/global")
	if len(model.state.VisibleItems) != 1 || model.state.VisibleItems[0].Source != claudefs.SkillSourceUser {
		t.Fatalf("global search mismatch: %+v", model.state.VisibleItems)
	}

	model.ApplyFilter("/cmd")
	if len(model.state.VisibleItems) != 1 || model.state.VisibleItems[0].Kind != claudefs.SkillKindCommand {
		t.Fatalf("cmd search mismatch: %+v", model.state.VisibleItems)
	}
}

func TestSkillListSortRebuildsCurrentView(t *testing.T) {
	model := &SkillListModel{
		state: DefaultResourceListState[claudefs.SkillResource]{
			Context: Context{Type: ContextAll},
			AllItems: []claudefs.SkillResource{
				{Name: "beta", Path: "/tmp/beta.md", Status: claudefs.SkillStatusInvalid},
				{Name: "alpha", Path: "/tmp/alpha.md", Status: claudefs.SkillStatusReady},
			},
		},
		sortBy:  SortBySkillName,
		sortAsc: true,
	}

	model.applyContext()
	model.sortBy = SortBySkillStatus
	model.sortAsc = true
	model.sortSkills(model.state.AllItems)
	model.applyContext()

	if len(model.state.VisibleItems) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(model.state.VisibleItems))
	}
	if model.state.VisibleItems[0].Status != claudefs.SkillStatusInvalid {
		t.Fatalf("expected invalid skill first after status sort, got %s", model.state.VisibleItems[0].Status)
	}
}

func TestNormalizeSkillSearchQuery(t *testing.T) {
	tests := map[string]string{
		"user":      "user",
		"/user":     "user",
		" /plugin ": "plugin",
		"  Ready  ": "ready",
	}

	for input, want := range tests {
		if got := normalizeResourceSearchQuery(input); got != want {
			t.Fatalf("normalizeResourceSearchQuery(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSkillSortCycleFollowsVisibleColumnOrder(t *testing.T) {
	model := &SkillListModel{sortBy: SortBySkillName}

	model.sortBy = (model.sortBy + 1) % 4
	if model.sortBy != SortBySkillType {
		t.Fatalf("first cycle = %v, want SortBySkillType", model.sortBy)
	}

	model.sortBy = (model.sortBy + 1) % 4
	if model.sortBy != SortBySkillStatus {
		t.Fatalf("second cycle = %v, want SortBySkillStatus", model.sortBy)
	}

	model.sortBy = (model.sortBy + 1) % 4
	if model.sortBy != SortBySkillScope {
		t.Fatalf("third cycle = %v, want SortBySkillScope", model.sortBy)
	}
}

func TestSkillListReloadPreservesFocusedSkill(t *testing.T) {
	model := &SkillListModel{
		state: DefaultResourceListState[claudefs.SkillResource]{
			Context: Context{Type: ContextAll},
			VisibleItems: []claudefs.SkillResource{
				{Name: "alpha", Path: "/tmp/alpha.md", Source: claudefs.SkillSourceUser},
				{Name: "beta", Path: "/tmp/beta.md", Source: claudefs.SkillSourceProject},
				{Name: "gamma", Path: "/tmp/gamma.md", Source: claudefs.SkillSourcePlugin},
			},
			Cursor: 1,
		},
	}

	model.captureCursorForReload()
	model.Update(skillsLoadedMsg{
		result: claudefs.SkillScanResult{
			Skills: []claudefs.SkillResource{
				{Name: "gamma", Path: "/tmp/gamma.md", Source: claudefs.SkillSourcePlugin},
				{Name: "beta", Path: "/tmp/beta.md", Source: claudefs.SkillSourceProject},
				{Name: "alpha", Path: "/tmp/alpha.md", Source: claudefs.SkillSourceUser},
			},
		},
	})

	if model.state.Cursor != 1 {
		t.Fatalf("cursor = %d, want 1", model.state.Cursor)
	}
	if model.state.VisibleItems[model.state.Cursor].Name != "beta" {
		t.Fatalf("focused skill = %s, want beta", model.state.VisibleItems[model.state.Cursor].Name)
	}
}

func TestSkillSetContextClearsActiveFilter(t *testing.T) {
	model := &SkillListModel{
		state: DefaultResourceListState[claudefs.SkillResource]{
			Context: Context{Type: ContextAll},
			ContextItems: []claudefs.SkillResource{
				{Name: "alpha", Source: claudefs.SkillSourceUser},
			},
			FilterQuery: "alpha",
		},
	}

	model.SetContext(Context{Type: ContextProject, Value: "cc9s"})

	if model.state.FilterQuery != "" {
		t.Fatalf("filterQuery = %q, want empty", model.state.FilterQuery)
	}
}
