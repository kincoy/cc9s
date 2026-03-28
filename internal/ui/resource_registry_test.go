package ui

import "testing"

func TestResourceRegistryFindByCommand(t *testing.T) {
	registry := newResourceRegistry()

	descriptor, ok := registry.FindByCommand("skills")
	if !ok {
		t.Fatal("expected skills descriptor")
	}
	if descriptor.Resource != ResourceSkills {
		t.Fatalf("resource = %v, want ResourceSkills", descriptor.Resource)
	}
	if !descriptor.Capabilities.SupportsEdit {
		t.Fatal("expected skills descriptor to advertise edit support")
	}
}

func TestResourceRegistryResolveSkillTargetContextFromProjectsDefaultsToAll(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceProjects)

	descriptor := app.resourceRegistry.MustGet(ResourceSkills)
	ctx := descriptor.ResolveTargetContext(app)
	if ctx.Type != ContextAll {
		t.Fatalf("context = %#v, want all context", ctx)
	}
}

func TestResourceRegistryResolveSessionTargetContextFromProjectsDefaultsToAll(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceProjects)

	descriptor := app.resourceRegistry.MustGet(ResourceSessions)
	ctx := descriptor.ResolveTargetContext(app)
	if ctx.Type != ContextAll {
		t.Fatalf("context = %#v, want all context", ctx)
	}
}

func TestResourceRegistryResolveAgentTargetContextFromSessionsPreservesProject(t *testing.T) {
	app := NewAppModel()
	app.setActiveResource(ResourceSessions)
	app.sessionList = NewSessionListModelForProject("cc9s")
	app.globalProjectContext = Context{Type: ContextProject, Value: "cc9s"}

	descriptor := app.resourceRegistry.MustGet(ResourceAgents)
	ctx := descriptor.ResolveTargetContext(app)
	if ctx.Type != ContextProject || ctx.Value != "cc9s" {
		t.Fatalf("context = %#v, want project context for cc9s", ctx)
	}
}

func TestSessionsDescriptorSupportsAllContextShortcut(t *testing.T) {
	registry := newResourceRegistry()

	descriptor := registry.MustGet(ResourceSessions)
	if !descriptor.Capabilities.SupportsAllContextShortcut {
		t.Fatal("expected sessions descriptor to support all-context shortcut")
	}

	projects := registry.MustGet(ResourceProjects)
	if projects.Capabilities.SupportsAllContextShortcut {
		t.Fatal("expected projects descriptor to not support all-context shortcut")
	}
}
