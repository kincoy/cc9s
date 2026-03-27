package ui

import (
	"strings"
	"testing"
	"time"

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

func TestSessionListViewShowsCleanupHintsColumn(t *testing.T) {
	model := &SessionListModel{
		context: Context{Type: ContextAll},
		sessions: []claudefs.GlobalSession{
			{
				ProjectName: "demo",
				Session: claudefs.Session{
					ID:           "stale-1",
					Summary:      "tiny stale session",
					LastActiveAt: time.Now().Add(-72 * time.Hour),
					EventCount:   2,
					FileSize:     500,
					Lifecycle: claudefs.SessionLifecycleSnapshot{
						State: claudefs.SessionLifecycleStale,
					},
				},
			},
		},
		contextSessions: []claudefs.GlobalSession{
			{
				ProjectName: "demo",
				Session: claudefs.Session{
					ID:           "stale-1",
					Summary:      "tiny stale session",
					LastActiveAt: time.Now().Add(-72 * time.Hour),
					EventCount:   2,
					FileSize:     500,
					Lifecycle: claudefs.SessionLifecycleSnapshot{
						State: claudefs.SessionLifecycleStale,
					},
				},
			},
		},
	}

	model.SetCleanupHints(true)
	out := model.View(140, 12)

	for _, want := range []string{"RECOMMEND", "Delete"} {
		if !strings.Contains(out, want) {
			t.Fatalf("cleanup hints output missing %q:\n%s", want, out)
		}
	}
}

func TestSessionListRecommendationSearch(t *testing.T) {
	model := &SessionListModel{}
	sessions := []claudefs.GlobalSession{
		{
			ProjectName: "demo",
			Session: claudefs.Session{
				ID:           "s-001",
				ProjectPath:  "/tmp/demo",
				LastActiveAt: time.Now().Add(-72 * time.Hour),
				EventCount:   2,
				FileSize:     500,
				Summary:      "hi",
				Lifecycle: claudefs.SessionLifecycleSnapshot{
					State: claudefs.SessionLifecycleStale,
				},
			},
		},
		{
			ProjectName: "demo",
			Session: claudefs.Session{
				ID:           "s-002",
				ProjectPath:  "/tmp/demo",
				LastActiveAt: time.Now().Add(-2 * time.Hour),
				EventCount:   300,
				FileSize:     200000,
				Summary:      "implement the full authentication system with JWT tokens",
				Lifecycle: claudefs.SessionLifecycleSnapshot{
					State: claudefs.SessionLifecycleCompleted,
				},
			},
		},
	}

	model.contextSessions = sessions

	model.ApplyFilter("del")
	if len(model.sessions) != 1 || model.sessions[0].Session.ID != "s-001" {
		t.Fatalf("delete recommendation search mismatch: %+v", model.sessions)
	}

	model.ApplyFilter("keep")
	if len(model.sessions) != 1 || model.sessions[0].Session.ID != "s-002" {
		t.Fatalf("keep recommendation search mismatch: %+v", model.sessions)
	}
}
