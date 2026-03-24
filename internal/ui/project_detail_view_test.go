package ui

import (
	"strings"
	"testing"

	"github.com/kincoy/cc9s/internal/claudefs"
)

func TestProjectDetailViewRendersMetadataAndResources(t *testing.T) {
	view := NewProjectDetailViewModel(claudefs.Project{
		Name:               "cc9s",
		Path:               "/tmp/cc9s",
		EncodedPath:        "tmp-cc9s",
		SessionCount:       12,
		ActiveSessionCount: 2,
		SkillCount:         3,
		CommandCount:       1,
		AgentCount:         2,
		HasSkillsRoot:      true,
		HasCommandsRoot:    true,
		HasAgentsRoot:      false,
	})

	rendered := view.ViewBox(100)
	for _, expected := range []string{
		"Project Details: cc9s",
		"/tmp/cc9s",
		"12 total / 2 active",
		"Local Skills:",
		"Local Commands:",
		"Local Agents:",
		"Context views include global user/plugin",
		"resources",
		"Present (3)",
		"Missing",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("expected rendered detail to contain %q, got:\n%s", expected, rendered)
		}
	}
}
