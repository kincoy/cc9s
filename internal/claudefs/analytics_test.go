package claudefs

import (
	"testing"
	"time"
)

func TestComputeProjectActivityScore(t *testing.T) {
	now := time.Now()
	proj := Project{Name: "test-proj"}
	sessions := []Session{
		{StartTime: now.AddDate(0, 0, -1), Lifecycle: SessionLifecycleSnapshot{State: SessionLifecycleActive}},
		{StartTime: now.AddDate(0, 0, -2), Lifecycle: SessionLifecycleSnapshot{State: SessionLifecycleIdle}},
		{StartTime: now.AddDate(0, 0, -3), Lifecycle: SessionLifecycleSnapshot{State: SessionLifecycleCompleted}},
	}

	score := computeProjectActivityScore(proj, sessions)

	if score < 0 || score > 100 {
		t.Fatalf("activity score out of range: %d", score)
	}
}

func TestComputeHealthMetricsSortsLowestProjectFirst(t *testing.T) {
	projects := []Project{
		{Name: "healthy-proj"},
		{Name: "unhealthy-proj"},
	}
	sessions := []GlobalSession{
		{
			Session: Session{
				ProjectPath: "/test/healthy-proj",
				Lifecycle:   SessionLifecycleSnapshot{State: SessionLifecycleCompleted},
			},
			ProjectName: "healthy-proj",
		},
		{
			Session: Session{
				ProjectPath: "/test/unhealthy-proj",
				Lifecycle:   SessionLifecycleSnapshot{State: SessionLifecycleStale},
			},
			ProjectName: "unhealthy-proj",
		},
	}

	result := computeHealthMetrics(projects, sessions)

	if result.EnvironmentScore < 0 || result.EnvironmentScore > 100 {
		t.Fatalf("environment score out of range: %d", result.EnvironmentScore)
	}
	if len(result.ProjectScores) != 2 {
		t.Fatalf("expected 2 project scores, got %d", len(result.ProjectScores))
	}
	if result.ProjectScores[0].ProjectName != "unhealthy-proj" {
		t.Fatalf("expected unhealthy project first, got %s", result.ProjectScores[0].ProjectName)
	}
	if len(result.Recommendations) == 0 {
		t.Fatal("expected at least one recommendation")
	}
}
