package ui

import (
	"testing"

	"github.com/kincoy/cc9s/internal/claudefs"
)

func TestSessionListLifecycleSearch(t *testing.T) {
	model := &SessionListModel{}
	sessions := []claudefs.GlobalSession{
		{
			ProjectName: "demo",
			Session: claudefs.Session{
				ID:          "active-1",
				ProjectPath: "/tmp/demo",
				Lifecycle: claudefs.SessionLifecycleSnapshot{
					State: claudefs.SessionLifecycleActive,
					Evidence: claudefs.StateEvidenceSummary{
						Reasons: []string{"Last activity was just now."},
					},
				},
			},
		},
		{
			ProjectName: "demo",
			Session: claudefs.Session{
				ID:          "idle-1",
				ProjectPath: "/tmp/demo",
				Lifecycle: claudefs.SessionLifecycleSnapshot{
					State: claudefs.SessionLifecycleIdle,
					Evidence: claudefs.StateEvidenceSummary{
						Reasons: []string{"Last activity was outside the active window."},
					},
				},
			},
		},
		{
			ProjectName: "demo",
			Session: claudefs.Session{
				ID:          "completed-1",
				ProjectPath: "/tmp/demo",
				Lifecycle: claudefs.SessionLifecycleSnapshot{
					State: claudefs.SessionLifecycleCompleted,
					Evidence: claudefs.StateEvidenceSummary{
						Reasons: []string{"Last activity was beyond the idle window."},
					},
				},
			},
		},
	}

	model.contextSessions = sessions

	model.ApplyFilter("active")
	if len(model.sessions) != 1 || model.sessions[0].Session.ID != "active-1" {
		t.Fatalf("active search mismatch: %+v", model.sessions)
	}

	model.ApplyFilter("ac")
	if len(model.sessions) != 1 || model.sessions[0].Session.ID != "active-1" {
		t.Fatalf("ac prefix search mismatch: %+v", model.sessions)
	}

	model.ApplyFilter("/completed")
	if len(model.sessions) != 1 || model.sessions[0].Session.ID != "completed-1" {
		t.Fatalf("/completed search mismatch: %+v", model.sessions)
	}
}

func TestNormalizeSearchQuery(t *testing.T) {
	tests := map[string]string{
		"active":      "active",
		"/active":     "active",
		" /stale  ":   "stale",
		" completed ": "completed",
	}

	for input, want := range tests {
		if got := normalizeSearchQuery(input); got != want {
			t.Errorf("normalizeSearchQuery(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestQueryLooksLikeLifecycleState(t *testing.T) {
	tests := map[string]bool{
		"a":         true,
		"ac":        true,
		"comp":      true,
		"sta":       true,
		"actx":      false,
		"proj":      false,
		"activity":  false,
		"completed": true,
	}

	for input, want := range tests {
		if got := queryLooksLikeLifecycleState(input); got != want {
			t.Errorf("queryLooksLikeLifecycleState(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestSessionListReloadPreservesFocusedSession(t *testing.T) {
	model := &SessionListModel{
		context: Context{Type: ContextAll},
		sessions: []claudefs.GlobalSession{
			{Session: claudefs.Session{ID: "one"}},
			{Session: claudefs.Session{ID: "two"}},
			{Session: claudefs.Session{ID: "three"}},
		},
		cursor: 1,
	}

	model.captureCursorForReload()

	model.Update(sessionsLoadedMsg{
		sessions: []claudefs.GlobalSession{
			{Session: claudefs.Session{ID: "zero"}},
			{Session: claudefs.Session{ID: "two"}},
			{Session: claudefs.Session{ID: "three"}},
		},
	})

	if model.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", model.cursor)
	}
	if model.sessions[model.cursor].Session.ID != "two" {
		t.Fatalf("focused session = %s, want two", model.sessions[model.cursor].Session.ID)
	}
}

func TestSessionListReloadFallsBackToPreviousIndex(t *testing.T) {
	model := &SessionListModel{
		context: Context{Type: ContextAll},
		sessions: []claudefs.GlobalSession{
			{Session: claudefs.Session{ID: "one"}},
			{Session: claudefs.Session{ID: "two"}},
			{Session: claudefs.Session{ID: "three"}},
		},
		cursor: 2,
	}

	model.captureCursorForReload()

	model.Update(sessionsLoadedMsg{
		sessions: []claudefs.GlobalSession{
			{Session: claudefs.Session{ID: "zero"}},
			{Session: claudefs.Session{ID: "one"}},
		},
	})

	if model.cursor != 1 {
		t.Fatalf("cursor = %d, want 1 after clamp", model.cursor)
	}
}
