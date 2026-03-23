package claudefs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClassifySessionLifecycle(t *testing.T) {
	now := time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		lastActiveAt    time.Time
		hasActiveMarker bool
		health          SessionHealth
		wantState       SessionLifecycleState
		wantReason      string
	}{
		{
			name:         "recent activity without marker stays active",
			lastActiveAt: now.Add(-5 * time.Minute),
			health:       SessionHealth{IsReliable: true},
			wantState:    SessionLifecycleActive,
			wantReason:   "inside the 15m active window",
		},
		{
			name:         "inactive within idle window becomes idle",
			lastActiveAt: now.Add(-2 * time.Hour),
			health:       SessionHealth{IsReliable: true},
			wantState:    SessionLifecycleIdle,
			wantReason:   "within the 24h idle window",
		},
		{
			name:         "inactive beyond idle window becomes completed",
			lastActiveAt: now.Add(-48 * time.Hour),
			health:       SessionHealth{IsReliable: true},
			wantState:    SessionLifecycleCompleted,
			wantReason:   "historical but still normal",
		},
		{
			name:            "marker older than active window but healthy becomes idle",
			lastActiveAt:    now.Add(-2 * time.Hour),
			hasActiveMarker: true,
			health:          SessionHealth{IsReliable: true},
			wantState:       SessionLifecycleIdle,
			wantReason:      "treated as idle, not stale",
		},
		{
			name:            "recent activity with coherent marker stays active",
			lastActiveAt:    now.Add(-2 * time.Minute),
			hasActiveMarker: true,
			health:          SessionHealth{IsReliable: true},
			wantState:       SessionLifecycleActive,
			wantReason:      "active marker is present and still coherent",
		},
		{
			name:         "unreliable session becomes stale",
			lastActiveAt: now.Add(-2 * time.Hour),
			health: SessionHealth{
				IsReliable: false,
				Problem:    "The session file only has residue events and no normal user/assistant session chain.",
			},
			wantState:  SessionLifecycleStale,
			wantReason: "no normal user/assistant session chain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifySessionLifecycle(tt.lastActiveAt, tt.hasActiveMarker, tt.health, now, DefaultActivityWindow)
			if got.State != tt.wantState {
				t.Fatalf("ClassifySessionLifecycle() state = %s, want %s", got.State, tt.wantState)
			}
			if len(got.Evidence.Reasons) < 2 || len(got.Evidence.Reasons) > 4 {
				t.Fatalf("expected 2-4 reasons, got %d: %v", len(got.Evidence.Reasons), got.Evidence.Reasons)
			}
			if !strings.Contains(strings.ToLower(strings.Join(got.Evidence.Reasons, " ")), strings.ToLower(tt.wantReason)) {
				t.Fatalf("expected reasons to mention %q, got %v", tt.wantReason, got.Evidence.Reasons)
			}
		})
	}
}

func TestClassifySessionLifecycleStaleIsSignalTrustIssue(t *testing.T) {
	now := time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC)

	completed := ClassifySessionLifecycle(now.Add(-72*time.Hour), false, SessionHealth{IsReliable: true}, now, DefaultActivityWindow)
	if completed.State != SessionLifecycleCompleted {
		t.Fatalf("old inactive session state = %s, want %s", completed.State, SessionLifecycleCompleted)
	}

	stale := ClassifySessionLifecycle(now.Add(-72*time.Hour), true, SessionHealth{IsReliable: false, Problem: "The session file only has residue events and no normal user/assistant session chain."}, now, DefaultActivityWindow)
	if stale.State != SessionLifecycleStale {
		t.Fatalf("marker conflict state = %s, want %s", stale.State, SessionLifecycleStale)
	}

	reasons := strings.ToLower(strings.Join(stale.Evidence.Reasons, " "))
	if !strings.Contains(reasons, "stale") && !strings.Contains(reasons, "trustworthy") {
		t.Fatalf("stale reasons should explain signal trust, got %v", stale.Evidence.Reasons)
	}
}

func TestInspectSessionFile(t *testing.T) {
	dir := t.TempDir()

	t.Run("healthy session", func(t *testing.T) {
		path := filepath.Join(dir, "healthy.jsonl")
		content := strings.Join([]string{
			`{"sessionId":"abc","cwd":"/tmp/demo","version":"1.0.0","type":"user","message":{"role":"user","content":"hello"},"timestamp":"2026-03-10T10:00:00Z"}`,
			`{"sessionId":"abc","cwd":"/tmp/demo","version":"1.0.0","type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"hi"}]},"timestamp":"2026-03-10T10:00:01Z"}`,
		}, "\n")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}

		got := inspectSessionFile(path)
		if !got.Health.IsReliable {
			t.Fatalf("healthy session marked unreliable: %+v", got.Health)
		}
		if got.Summary != "hello" {
			t.Fatalf("summary = %q, want hello", got.Summary)
		}
	})

	t.Run("snapshot residue is stale", func(t *testing.T) {
		path := filepath.Join(dir, "stale.jsonl")
		content := strings.Join([]string{
			`{"type":"file-history-snapshot","snapshot":{"timestamp":"2026-03-10T03:40:59.640Z"}}`,
			`{"type":"file-history-snapshot","snapshot":{"timestamp":"2026-03-10T03:43:01.744Z"}}`,
		}, "\n")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}

		got := inspectSessionFile(path)
		if got.Health.IsReliable {
			t.Fatalf("snapshot residue should be unreliable: %+v", got.Health)
		}
		if !strings.Contains(strings.ToLower(got.Health.Problem), "no normal user/assistant") {
			t.Fatalf("unexpected health problem: %q", got.Health.Problem)
		}
	})
}

func TestLifecycleStateMatchesQuery(t *testing.T) {
	tests := []struct {
		state SessionLifecycleState
		query string
		want  bool
	}{
		{SessionLifecycleActive, "active", true},
		{SessionLifecycleIdle, "idl", true},
		{SessionLifecycleCompleted, "completed", true},
		{SessionLifecycleStale, "stale", true},
		{SessionLifecycleCompleted, "active", false},
	}

	for _, tt := range tests {
		if got := LifecycleStateMatchesQuery(tt.state, tt.query); got != tt.want {
			t.Errorf("LifecycleStateMatchesQuery(%s, %q) = %v, want %v", tt.state, tt.query, got, tt.want)
		}
	}
}

func TestSummarizeLifecycleSessions(t *testing.T) {
	sessions := []Session{
		{Lifecycle: SessionLifecycleSnapshot{State: SessionLifecycleActive}},
		{Lifecycle: SessionLifecycleSnapshot{State: SessionLifecycleIdle}},
		{Lifecycle: SessionLifecycleSnapshot{State: SessionLifecycleCompleted}},
		{Lifecycle: SessionLifecycleSnapshot{State: SessionLifecycleCompleted}},
		{Lifecycle: SessionLifecycleSnapshot{State: SessionLifecycleStale}},
	}

	got := SummarizeLifecycleSessions(sessions)
	if got.Total != 5 || got.Active != 1 || got.Idle != 1 || got.Completed != 2 || got.Stale != 1 {
		t.Fatalf("unexpected lifecycle summary: %+v", got)
	}
}
