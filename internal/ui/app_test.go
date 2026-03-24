package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/kincoy/cc9s/internal/claudefs"
)

func TestSwitchToSkillsPreservesProjectContextFromSessions(t *testing.T) {
	app := NewAppModel()
	app.currentScreen = ScreenSessions
	app.currentResource = ResourceSessions
	app.sessionList = NewSessionListModelForProject("cc9s")

	model, cmd := app.Update(SwitchResourceMsg{Resource: ResourceSkills})
	if cmd == nil {
		t.Fatal("expected switch-to-skills command")
	}

	appModel, ok := model.(*AppModel)
	if !ok {
		t.Fatalf("expected *AppModel, got %T", model)
	}

	if appModel.currentScreen != ScreenSkills {
		t.Fatalf("screen = %v, want ScreenSkills", appModel.currentScreen)
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
	app.currentScreen = ScreenProjects
	app.currentResource = ResourceProjects

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
	app.currentScreen = ScreenSkills
	app.currentResource = ResourceSkills
	app.skillList = NewSkillListModel()
	app.skillList.loading = false
	app.skillList.contextSkills = []claudefs.SkillResource{
		{Name: "alpha", Source: claudefs.SkillSourceUser},
		{Name: "beta", Source: claudefs.SkillSourcePlugin},
	}
	app.skillList.ApplyFilter("plugin")

	if len(app.skillList.skills) != 1 {
		t.Fatalf("expected filtered list before esc, got %d items", len(app.skillList.skills))
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
	if len(appModel.skillList.skills) != 2 {
		t.Fatalf("expected full list after esc clear, got %d items", len(appModel.skillList.skills))
	}
}
