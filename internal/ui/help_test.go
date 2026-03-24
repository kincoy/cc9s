package ui

import (
	"strings"
	"testing"
)

func TestRenderHelpUsesRegistryCommandNamesAndActiveCapabilities(t *testing.T) {
	registry := newResourceRegistry()

	output := renderHelp(120, 40, registry, registry.MustGet(ResourceSkills))
	if !strings.Contains(output, ":skills") {
		t.Fatal("expected help to include skills command from registry")
	}
	if !strings.Contains(output, ":agents") {
		t.Fatal("expected help to include agents command from registry")
	}
	if !strings.Contains(output, "Switch to all projects") {
		t.Fatal("expected skills help to include all-context shortcut guidance")
	}

	output = renderHelp(120, 40, registry, registry.MustGet(ResourceProjects))
	if strings.Contains(output, "Switch to all projects") {
		t.Fatal("expected projects help to omit active all-context shortcut guidance")
	}
}
