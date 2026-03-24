package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/kincoy/cc9s/internal/claudefs"
)

func TestProjectListDetailShortcut(t *testing.T) {
	model := &ProjectListModel{
		projects: []claudefs.Project{
			{Name: "cc9s", Path: "/tmp/cc9s"},
		},
	}

	cmd := model.Update(tea.KeyPressMsg{Text: "d", Code: 'd'})
	if cmd == nil {
		t.Fatal("expected detail command")
	}
	msg := cmd()
	if _, ok := msg.(ShowProjectDetailMsg); !ok {
		t.Fatalf("expected ShowProjectDetailMsg, got %T", msg)
	}
}

func TestProjectSortShortcutFollowsColumnOrder(t *testing.T) {
	model := NewProjectListModel()
	model.lastWidth = 160

	expected := []SortField{
		SortBySize,
		SortByName,
		SortByPath,
		SortBySessionCount,
		SortBySkillCount,
		SortByAgentCount,
		SortByActivity,
	}

	for i, want := range expected {
		model.Update(tea.KeyPressMsg{Text: "s", Code: 's'})
		if model.sortBy != want {
			t.Fatalf("step %d: sortBy = %v, want %v", i, model.sortBy, want)
		}
	}
}

func TestProjectSortShortcutSkipsPathWhenColumnHidden(t *testing.T) {
	model := NewProjectListModel()
	model.lastWidth = 120

	expected := []SortField{
		SortBySize,
		SortByName,
		SortBySessionCount,
		SortBySkillCount,
		SortByAgentCount,
		SortByActivity,
	}

	for i, want := range expected {
		model.Update(tea.KeyPressMsg{Text: "s", Code: 's'})
		if model.sortBy != want {
			t.Fatalf("step %d: sortBy = %v, want %v", i, model.sortBy, want)
		}
	}
}

func TestProjectSortByLocalSkillsUsesSkillsAndCommands(t *testing.T) {
	model := &ProjectListModel{
		allProjects: []claudefs.Project{
			{Name: "small", SkillCount: 1, CommandCount: 0},
			{Name: "large", SkillCount: 2, CommandCount: 3},
		},
		sortBy:  SortBySkillCount,
		sortAsc: false,
	}

	model.sortProjects()

	if got := model.allProjects[0].Name; got != "large" {
		t.Fatalf("first project = %q, want large", got)
	}
}

func TestProjectsLoadedAppliesCurrentSortAndFilter(t *testing.T) {
	model := &ProjectListModel{
		sortBy:      SortByName,
		sortAsc:     true,
		filterQuery: "a",
	}

	model.Update(projectsLoadedMsg{
		result: claudefs.ScanResult{
			Projects: []claudefs.Project{
				{Name: "zeta", Path: "/tmp/zeta"},
				{Name: "alpha", Path: "/tmp/alpha"},
				{Name: "beta", Path: "/tmp/beta"},
			},
		},
	})

	if len(model.projects) != 3 {
		t.Fatalf("filtered projects = %d, want 3", len(model.projects))
	}
	if model.projects[0].Name != "alpha" || model.projects[1].Name != "beta" || model.projects[2].Name != "zeta" {
		t.Fatalf("projects not sorted after load: %#v", model.projects)
	}
}
