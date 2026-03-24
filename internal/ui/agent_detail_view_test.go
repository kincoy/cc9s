package ui

import (
	"strings"
	"testing"

	"github.com/kincoy/cc9s/internal/claudefs"
)

func TestAgentDetailViewRendersMetadataAndReasons(t *testing.T) {
	view := NewAgentDetailViewModel(claudefs.AgentResource{
		Name:              "go-reviewer",
		Path:              "/tmp/go-reviewer.md",
		Source:            claudefs.AgentSourcePlugin,
		Scope:             claudefs.AgentScopePlugin,
		PluginName:        "everything-claude-code",
		Model:             "sonnet",
		Tools:             []string{"Read", "Grep"},
		Status:            claudefs.AgentStatusInvalid,
		Description:       "Reviews Go code.",
		ValidationReasons: []string{"Agent file is not recognized by Claude Code"},
	})

	rendered := view.ViewBox(100)
	for _, expected := range []string{
		"Agent Details: go-reviewer",
		"everything-claude-code",
		"sonnet",
		"Read, Grep",
		"Agent file is not recognized by Claude Code",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("expected rendered detail to contain %q, got:\n%s", expected, rendered)
		}
	}
}
