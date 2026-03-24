package ui

import (
	"testing"

	"github.com/kincoy/cc9s/internal/claudefs"
)

func TestSkillListSearchNormalizesSlashPrefix(t *testing.T) {
	model := &SkillListModel{
		contextSkills: []claudefs.SkillResource{
			{Name: "alpha", Source: claudefs.SkillSourceUser, Scope: claudefs.SkillScopeUser},
			{Name: "beta", Source: claudefs.SkillSourceProject, Scope: claudefs.SkillScopeProject},
		},
	}

	model.ApplyFilter("/user")

	if len(model.skills) != 1 {
		t.Fatalf("expected 1 filtered skill, got %d", len(model.skills))
	}
	if model.skills[0].Source != claudefs.SkillSourceUser {
		t.Fatalf("expected user skill, got %s", model.skills[0].Source)
	}
}

func TestSkillListSearchMatchesVisibleAliases(t *testing.T) {
	model := &SkillListModel{
		contextSkills: []claudefs.SkillResource{
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
	}

	model.ApplyFilter("/global")
	if len(model.skills) != 1 || model.skills[0].Source != claudefs.SkillSourceUser {
		t.Fatalf("global search mismatch: %+v", model.skills)
	}

	model.ApplyFilter("/cmd")
	if len(model.skills) != 1 || model.skills[0].Kind != claudefs.SkillKindCommand {
		t.Fatalf("cmd search mismatch: %+v", model.skills)
	}
}

func TestSkillListSortRebuildsCurrentView(t *testing.T) {
	model := &SkillListModel{
		context: Context{Type: ContextAll},
		allSkills: []claudefs.SkillResource{
			{Name: "beta", Path: "/tmp/beta.md", Status: claudefs.SkillStatusInvalid},
			{Name: "alpha", Path: "/tmp/alpha.md", Status: claudefs.SkillStatusReady},
		},
		sortBy:  SortBySkillName,
		sortAsc: true,
	}

	model.applyContext()
	model.sortBy = SortBySkillStatus
	model.sortAsc = true
	model.sortSkills(model.allSkills)
	model.applyContext()

	if len(model.skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(model.skills))
	}
	if model.skills[0].Status != claudefs.SkillStatusInvalid {
		t.Fatalf("expected invalid skill first after status sort, got %s", model.skills[0].Status)
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
