package claudefs

import (
	"testing"
	"time"
)

func TestAssessSessionValue_StaleAlwaysRecommendDelete(t *testing.T) {
	stats := SessionStats{
		TurnCount:     1,
		ToolCallCount: 0,
		InputTokens:   100,
		OutputTokens:  50,
	}
	lifecycle := SessionLifecycleSnapshot{
		State: SessionLifecycleStale,
	}

	a := AssessSessionValue(stats, lifecycle, 100)

	if a.Recommendation != RecommendDelete {
		t.Errorf("Stale session: want Delete, got %s", a.Recommendation)
	}
}

func TestAssessSessionValue_HighValueSession(t *testing.T) {
	stats := SessionStats{
		TurnCount:     25,
		ToolCallCount: 40,
		InputTokens:   50000,
		OutputTokens:  30000,
		Duration:      2 * time.Hour,
	}
	lifecycle := SessionLifecycleSnapshot{
		State: SessionLifecycleCompleted,
	}

	a := AssessSessionValue(stats, lifecycle, 5000)

	if a.Recommendation != RecommendKeep {
		t.Errorf("High-value session: want Keep, got %s (score=%d)", a.Recommendation, a.Score)
	}
	if a.Score < 60 {
		t.Errorf("High-value session score too low: %d", a.Score)
	}
}

func TestAssessSessionValue_LowValueSession(t *testing.T) {
	stats := SessionStats{
		TurnCount:     2,
		ToolCallCount: 0,
		InputTokens:   200,
		OutputTokens:  100,
		Duration:      30 * time.Second,
	}
	lifecycle := SessionLifecycleSnapshot{
		State: SessionLifecycleCompleted,
	}

	a := AssessSessionValue(stats, lifecycle, 500)

	if a.Recommendation != RecommendDelete {
		t.Errorf("Low-value session: want Delete, got %s (score=%d)", a.Recommendation, a.Score)
	}
}

func TestAssessSessionValue_MediumValueSession(t *testing.T) {
	stats := SessionStats{
		TurnCount:     8,
		ToolCallCount: 5,
		InputTokens:   5000,
		OutputTokens:  3000,
		Duration:      15 * time.Minute,
	}
	lifecycle := SessionLifecycleSnapshot{
		State: SessionLifecycleCompleted,
	}

	a := AssessSessionValue(stats, lifecycle, 2000)

	if a.Recommendation != RecommendMaybe {
		t.Errorf("Medium-value session: want Maybe, got %s (score=%d)", a.Recommendation, a.Score)
	}
}

func TestAssessSessionValue_ActiveAlwaysKeep(t *testing.T) {
	stats := SessionStats{
		TurnCount:     1,
		ToolCallCount: 0,
	}
	lifecycle := SessionLifecycleSnapshot{
		State: SessionLifecycleActive,
	}

	a := AssessSessionValue(stats, lifecycle, 100)

	if a.Recommendation != RecommendKeep {
		t.Errorf("Active session: want Keep, got %s", a.Recommendation)
	}
}

func TestAssessSessionValue_IdleAlwaysKeep(t *testing.T) {
	stats := SessionStats{
		TurnCount:     3,
		ToolCallCount: 1,
	}
	lifecycle := SessionLifecycleSnapshot{
		State: SessionLifecycleIdle,
	}

	a := AssessSessionValue(stats, lifecycle, 800)

	if a.Recommendation != RecommendKeep {
		t.Errorf("Idle session: want Keep, got %s", a.Recommendation)
	}
}

func TestAssessSessionValue_Reasons(t *testing.T) {
	stats := SessionStats{
		TurnCount:     1,
		ToolCallCount: 0,
		InputTokens:   100,
		OutputTokens:  50,
	}
	lifecycle := SessionLifecycleSnapshot{
		State: SessionLifecycleCompleted,
	}

	a := AssessSessionValue(stats, lifecycle, 300)

	if len(a.Reasons) == 0 {
		t.Error("Assessment should have at least one reason")
	}
}

func TestScoreToRecommendation(t *testing.T) {
	tests := []struct {
		score int
		want  CleanupRecommendation
	}{
		{0, RecommendDelete},
		{29, RecommendDelete},
		{30, RecommendMaybe},
		{59, RecommendMaybe},
		{60, RecommendKeep},
		{100, RecommendKeep},
	}
	for _, tt := range tests {
		got := scoreToRecommendation(tt.score)
		if got != tt.want {
			t.Errorf("scoreToRecommendation(%d): want %s, got %s", tt.score, tt.want, got)
		}
	}
}

func TestQuickAssessSession_StaleSmallFile(t *testing.T) {
	s := Session{
		FileSize:   300,
		EventCount: 5,
		Lifecycle: SessionLifecycleSnapshot{
			State: SessionLifecycleStale,
		},
	}

	a := QuickAssessSession(s)

	if a.Recommendation != RecommendDelete {
		t.Errorf("Stale small file: want Delete, got %s", a.Recommendation)
	}
}

func TestQuickAssessSession_CompletedLargeFile(t *testing.T) {
	s := Session{
		FileSize:   200000,
		EventCount: 500,
		Summary:    "implement the full authentication system with JWT tokens",
		Lifecycle: SessionLifecycleSnapshot{
			State: SessionLifecycleCompleted,
		},
	}

	a := QuickAssessSession(s)

	if a.Recommendation != RecommendKeep {
		t.Errorf("Large completed session: want Keep, got %s (score=%d)", a.Recommendation, a.Score)
	}
}

func TestQuickAssessSession_CompletedTinySession(t *testing.T) {
	s := Session{
		FileSize:   800,
		EventCount: 3,
		Summary:    "hi",
		Lifecycle: SessionLifecycleSnapshot{
			State: SessionLifecycleCompleted,
		},
	}

	a := QuickAssessSession(s)

	if a.Recommendation != RecommendDelete {
		t.Errorf("Tiny completed session: want Delete, got %s (score=%d)", a.Recommendation, a.Score)
	}
}
