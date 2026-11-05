package ui

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/kincoy/cc9s/internal/claudefs"
)

func TestRenderSessionTableFitsWidth(t *testing.T) {
	sessions := []claudefs.GlobalSession{
		{
			ProjectName: "demo-project",
			Session: claudefs.Session{
				ID:           "d39f7a9f-327f-49fb-8c90-c653e193fa79",
				Name:         "邮件二期",
				Summary:      "implement the full authentication system with JWT tokens",
				LastActiveAt: time.Now(),
				EventCount:   120,
				FileSize:     2048,
				Lifecycle: claudefs.SessionLifecycleSnapshot{
					State: claudefs.SessionLifecycleCompleted,
				},
			},
		},
	}

	tests := []struct {
		width           int
		showProjectCol  bool
		showCleanupHint bool
	}{
		{width: 80, showProjectCol: false},
		{width: 80, showProjectCol: true},
		{width: 100, showProjectCol: true},
		{width: 120, showProjectCol: true},
		{width: 130, showProjectCol: true},
		{width: 160, showProjectCol: true, showCleanupHint: true},
	}

	for _, tt := range tests {
		out := renderSessionTable(
			sessions, 0, tt.width, 10, nil,
			tt.showProjectCol, SortBySessionActivity, false,
			"All Projects", tt.showCleanupHint,
		)
		for _, line := range strings.Split(strings.TrimSuffix(out, "\n"), "\n") {
			if w := lipgloss.Width(line); w > tt.width {
				t.Fatalf("width %d (project=%v): line width %d overflows:\n%s",
					tt.width, tt.showProjectCol, w, line)
			}
		}
		if !strings.Contains(out, "NAME") {
			t.Fatalf("width %d: missing NAME header:\n%s", tt.width, out)
		}
		if !strings.Contains(out, "邮件二期") {
			t.Fatalf("width %d: missing session name:\n%s", tt.width, out)
		}
	}
}
