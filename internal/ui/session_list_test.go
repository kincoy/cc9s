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
	model.viewport = NewViewportWithSize(80, 20)

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
	model.viewport = NewViewportWithSize(80, 20)

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
	model.viewport = NewViewportWithSize(140, 12)

	model.SetCleanupHints(true)
	out := model.View(140, 12)

	for _, want := range []string{"RECOMMEND", "Delete"} {
		if !strings.Contains(out, want) {
			t.Fatalf("cleanup hints output missing %q:\n%s", want, out)
		}
	}
}

func TestSessionListSearchByName(t *testing.T) {
	model := &SessionListModel{}
	model.contextSessions = []claudefs.GlobalSession{
		{ProjectName: "demo", Session: claudefs.Session{ID: "s-001", Name: "邮件二期"}},
		{ProjectName: "demo", Session: claudefs.Session{ID: "s-002", Name: "收入认领三期"}},
	}

	model.ApplyFilter("邮件")
	if len(model.sessions) != 1 || model.sessions[0].Session.ID != "s-001" {
		t.Fatalf("name search mismatch: %+v", model.sessions)
	}

	model.ApplyFilter("认领")
	if len(model.sessions) != 1 || model.sessions[0].Session.ID != "s-002" {
		t.Fatalf("name search mismatch: %+v", model.sessions)
	}
}

func TestSessionListSortByName(t *testing.T) {
	model := &SessionListModel{sortBy: SortBySessionName, sortAsc: true}
	sessions := []claudefs.GlobalSession{
		{Session: claudefs.Session{ID: "b", Name: "zeta"}},
		{Session: claudefs.Session{ID: "a", Name: "Alpha"}},
		{Session: claudefs.Session{ID: "c", Name: "beta"}},
	}

	model.sortGlobalSessions(sessions)

	want := []string{"Alpha", "beta", "zeta"}
	for i, name := range want {
		if sessions[i].Session.Name != name {
			t.Fatalf("position %d = %q, want %q", i, sessions[i].Session.Name, name)
		}
	}
}

func TestSessionListViewShowsNameColumn(t *testing.T) {
	model := &SessionListModel{
		context: Context{Type: ContextProject, Value: "demo"},
		sessions: []claudefs.GlobalSession{
			{
				ProjectName: "demo",
				Session: claudefs.Session{
					ID:           "d39f7a9f-327f-49fb-8c90-c653e193fa79",
					Name:         "邮件二期",
					Summary:      "implement the full authentication system with JWT tokens",
					LastActiveAt: time.Now(),
					EventCount:   12,
					FileSize:     2048,
					Lifecycle: claudefs.SessionLifecycleSnapshot{
						State: claudefs.SessionLifecycleCompleted,
					},
				},
			},
		},
	}
	model.contextSessions = model.sessions
	model.viewport = NewViewportWithSize(140, 12)
	model.updateViewportContent()

	out := model.View(140, 12)
	for _, want := range []string{"SESSION ID", "NAME", "邮件二期"} {
		if !strings.Contains(out, want) {
			t.Fatalf("session list output missing %q:\n%s", want, out)
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
