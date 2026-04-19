package ui

import (
	"strings"
	"testing"
)

func TestRenderHelpUsesRegistryCommandNamesAndActiveCapabilities(t *testing.T) {
	registry := newResourceRegistry()

	output := renderHelp(120, 120, registry, registry.MustGet(ResourceSkills), 0)
	if !strings.Contains(output, ":skills") {
		t.Fatal("expected help to include skills command from registry")
	}
	if !strings.Contains(output, ":agents") {
		t.Fatal("expected help to include agents command from registry")
	}
	if !strings.Contains(output, "Switch to all projects") {
		t.Fatal("expected skills help to include all-context shortcut guidance")
	}
	if !strings.Contains(output, ":health") {
		t.Fatal("expected global help to include health command guidance")
	}
	if !strings.Contains(output, ":cleanup") {
		t.Fatal("expected global help to include cleanup command guidance")
	}

	output = renderHelp(120, 120, registry, registry.MustGet(ResourceProjects), 0)
	if !strings.Contains(output, "Current resource: Projects") {
		t.Fatal("expected help to identify current projects resource")
	}
	if !strings.Contains(output, ":health") {
		t.Fatal("expected projects help to include health command guidance")
	}
	if !strings.Contains(output, "Toggle HEALTH column") {
		t.Fatal("expected projects help to describe health column toggle")
	}
	if !strings.Contains(output, ":cleanup") {
		t.Fatal("expected projects help to include cleanup command guidance")
	}
	if !strings.Contains(output, "Toggle STATUS column") {
		t.Fatal("expected projects help to describe cleanup STATUS column")
	}
	if !strings.Contains(output, "Switch to all projects") {
		t.Fatal("expected projects help to include all-context shortcut guidance")
	}

	output = renderHelp(120, 120, registry, registry.MustGet(ResourceSessions), 0)
	if !strings.Contains(output, "Current resource: Sessions") {
		t.Fatal("expected help to identify current sessions resource")
	}
	if !strings.Contains(output, ":health") {
		t.Fatal("expected sessions help to include health command guidance")
	}
	if !strings.Contains(output, ":cleanup") {
		t.Fatal("expected sessions help to include cleanup command guidance")
	}
	if !strings.Contains(output, "Toggle cleanup recommendations") {
		t.Fatal("expected sessions help to describe cleanup recommendations")
	}
	if !strings.Contains(output, "Switch to all projects") {
		t.Fatal("expected sessions help to include all-context shortcut guidance")
	}
}

func TestBuildHelpLinesUseStableGlobalSectionOrder(t *testing.T) {
	registry := newResourceRegistry()

	lines := buildHelpLines(registry, registry.MustGet(ResourceSessions))
	output := strings.Join(lines, "\n")

	expectedOrder := []string{
		"General",
		"Navigation",
		"Sorting",
		"Project Operations",
		"Session Operations",
		"Skill Operations",
		"Agent Operations",
		"Context",
		"Dialog",
	}
	lastIdx := -1
	for _, section := range expectedOrder {
		idx := strings.Index(output, section)
		if idx == -1 {
			t.Fatalf("expected help output to include %q section", section)
		}
		if idx < lastIdx {
			t.Fatalf("expected %q section to appear after previous section", section)
		}
		lastIdx = idx
	}
}
