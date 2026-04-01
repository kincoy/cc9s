package contract

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kincoy/cc9s/internal/claudefs"
)

func TestStatusResultJSONPreservesHealthField(t *testing.T) {
	data, err := json.Marshal(StatusResult{
		Projects:  2,
		Sessions:  8,
		Resources: 3,
		Health: &claudefs.HealthMetrics{
			EnvironmentScore: 72,
			ProjectScores: []claudefs.ProjectHealthScore{
				{ProjectName: "cc9s", HealthScore: 72, SessionCount: 4},
			},
			Recommendations: []string{"Environment health looks good"},
		},
	})
	if err != nil {
		t.Fatalf("marshal status result: %v", err)
	}

	out := string(data)
	if !strings.Contains(out, `"health":`) {
		t.Fatalf("status json missing health field:\n%s", out)
	}
	legacyField := `"dash` + `board":`
	if strings.Contains(out, legacyField) {
		t.Fatalf("status json should not include unexpected legacy field:\n%s", out)
	}
}

func TestErrorPayloadJSONShape(t *testing.T) {
	data, err := json.Marshal(ErrorPayload{Error: "boom"})
	if err != nil {
		t.Fatalf("marshal error payload: %v", err)
	}

	if got, want := string(data), `{"error":"boom"}`; got != want {
		t.Fatalf("error payload json = %s, want %s", got, want)
	}
}

func TestCleanupResultJSONIncludesAssessmentFields(t *testing.T) {
	data, err := json.Marshal(CleanupResult{
		DryRun: true,
		Summary: CleanupSummary{
			MatchedSessions: 3,
			MatchedProjects: 1,
			TotalSizeBytes:  4096,
			DeleteCount:     1,
			MaybeCount:      1,
			KeepCount:       1,
		},
		Sessions: []CleanupSessionMatch{{
			ID:             "session-1",
			Project:        "cc9s",
			State:          "Completed",
			AgeHours:       72,
			UpdatedAt:      "2026-03-25T10:00:00Z",
			Recommendation: "Delete",
			Score:          5,
			Reasons:        []string{"small session file"},
		}},
	})
	if err != nil {
		t.Fatalf("marshal cleanup result: %v", err)
	}

	out := string(data)
	for _, want := range []string{
		`"delete_count":1`,
		`"maybe_count":1`,
		`"keep_count":1`,
		`"recommendation":"Delete"`,
		`"score":5`,
		`"reasons":["small session file"]`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("cleanup json missing %q:\n%s", want, out)
		}
	}
}
