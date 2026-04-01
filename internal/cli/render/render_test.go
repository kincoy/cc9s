package render

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/cli/contract"
)

func TestWriteErrorJSONUsesStdoutPayload(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := WriteError(&stdout, &stderr, contract.OutputJSON, errors.New("boom")); err != nil {
		t.Fatalf("WriteError returned error: %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != `{"error":"boom"}` {
		t.Fatalf("stdout = %q, want JSON error payload", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty in json mode, got %q", stderr.String())
	}
}

func TestWriteErrorTextUsesStderr(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := WriteError(&stdout, &stderr, contract.OutputText, errors.New("boom")); err != nil {
		t.Fatalf("WriteError returned error: %v", err)
	}

	if stdout.Len() != 0 {
		t.Fatalf("stdout should be empty in text mode, got %q", stdout.String())
	}
	if got := stderr.String(); got != "error: boom\n" {
		t.Fatalf("stderr = %q, want text error output", got)
	}
}

func TestWriteResultJSONRendersArrayForProjectList(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	result := contract.ProjectListResult{
		Projects: []contract.ProjectListEntry{
			{
				Name:               "cc9s",
				SessionCount:       3,
				ActiveSessionCount: 1,
				LastActiveAt:       "2026-03-25T10:00:00Z",
				SkillCount:         2,
				CommandCount:       1,
				AgentCount:         4,
				TotalSizeBytes:     2048,
				Path:               "/tmp/cc9s",
			},
		},
	}

	if err := WriteResult(&stdout, &stderr, contract.OutputJSON, result); err != nil {
		t.Fatalf("WriteResult returned error: %v", err)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty in json mode, got %q", stderr.String())
	}

	var payload []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("stdout is not a JSON array: %v\n%s", err, stdout.String())
	}
	if len(payload) != 1 {
		t.Fatalf("json array length = %d, want 1", len(payload))
	}
	if got := payload[0]["path"]; got != "/tmp/cc9s" {
		t.Fatalf("array item path = %v, want /tmp/cc9s", got)
	}
}

func TestWriteResultJSONRendersObjectForProjectDetail(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	result := contract.ProjectDetailResult{
		ProjectDetail: contract.ProjectDetail{
			Name:           "cc9s",
			Path:           "/tmp/cc9s",
			ClaudeRoot:     "/tmp/cc9s/.claude",
			LastActiveAt:   "2026-03-25T10:00:00Z",
			TotalSizeBytes: 4096,
			Sessions: contract.ProjectSessions{
				Total:  3,
				Active: 1,
			},
			Resources: contract.ProjectResources{
				Skills:   2,
				Commands: 1,
				Agents:   4,
			},
		},
	}

	if err := WriteResult(&stdout, &stderr, contract.OutputJSON, result); err != nil {
		t.Fatalf("WriteResult returned error: %v", err)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty in json mode, got %q", stderr.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("stdout is not a JSON object: %v\n%s", err, stdout.String())
	}
	if got := payload["path"]; got != "/tmp/cc9s" {
		t.Fatalf("object field path = %v, want /tmp/cc9s", got)
	}
}

func TestWriteResultJSONOmitsValidationReasonsForValidSkill(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	result := contract.SkillListResult{
		Skills: []contract.SkillListEntry{
			{
				Name:    "brainstorming",
				Type:    "skill",
				Scope:   "user",
				Status:  "valid",
				Project: "",
				Path:    "/tmp/skills/brainstorming",
			},
			{
				Name:              "broken",
				Type:              "skill",
				Scope:             "project",
				Status:            "invalid",
				Project:           "cc9s",
				Path:              "/tmp/cc9s/.claude/skills/broken",
				ValidationReasons: []string{"missing metadata"},
			},
		},
	}

	if err := WriteResult(&stdout, &stderr, contract.OutputJSON, result); err != nil {
		t.Fatalf("WriteResult returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty in json mode, got %q", stderr.String())
	}

	var payload []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("stdout is not a JSON array: %v\n%s", err, stdout.String())
	}
	if len(payload) != 2 {
		t.Fatalf("json array length = %d, want 2", len(payload))
	}
	if _, ok := payload[0]["validation_reasons"]; ok {
		t.Fatalf("valid skill unexpectedly included validation_reasons: %v", payload[0]["validation_reasons"])
	}
	if got := payload[1]["validation_reasons"]; got == nil {
		t.Fatalf("invalid skill should include validation_reasons")
	}
}

func TestWriteResultJSONOmitsValidationReasonsForExplicitEmptySlices(t *testing.T) {
	t.Run("skills", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		result := contract.SkillListResult{
			Skills: []contract.SkillListEntry{
				{
					Name:   "compact",
					Type:   "skill",
					Scope:  "user",
					Status: "valid",
					Path:   "/tmp/skills/compact",
				},
				{
					Name:              "explicit-empty",
					Type:              "skill",
					Scope:             "project",
					Status:            "invalid",
					Project:           "cc9s",
					Path:              "/tmp/cc9s/.claude/skills/explicit-empty",
					ValidationReasons: []string{},
				},
			},
		}

		if err := WriteResult(&stdout, &stderr, contract.OutputJSON, result); err != nil {
			t.Fatalf("WriteResult returned error: %v", err)
		}

		var payload []map[string]any
		if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
			t.Fatalf("stdout is not a JSON array: %v\n%s", err, stdout.String())
		}

		if _, ok := payload[0]["validation_reasons"]; ok {
			t.Fatalf("nil validation_reasons entry should stay compact, got %v", payload[0])
		}
		if _, ok := payload[1]["validation_reasons"]; ok {
			t.Fatalf("explicit empty validation_reasons should still serialize compactly, got %v", payload[1])
		}
	})

	t.Run("agents", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		result := contract.AgentListResult{
			Agents: []contract.AgentListEntry{
				{
					Name:   "compact-agent",
					Scope:  "project",
					Status: "valid",
					Path:   "/tmp/agents/compact-agent.md",
				},
				{
					Name:              "explicit-empty-agent",
					Scope:             "project",
					Status:            "invalid",
					Project:           "cc9s",
					Path:              "/tmp/agents/explicit-empty-agent.md",
					ValidationReasons: []string{},
				},
			},
		}

		if err := WriteResult(&stdout, &stderr, contract.OutputJSON, result); err != nil {
			t.Fatalf("WriteResult returned error: %v", err)
		}

		var payload []map[string]any
		if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
			t.Fatalf("stdout is not a JSON array: %v\n%s", err, stdout.String())
		}

		if _, ok := payload[0]["validation_reasons"]; ok {
			t.Fatalf("nil validation_reasons entry should stay compact, got %v", payload[0])
		}
		if _, ok := payload[1]["validation_reasons"]; ok {
			t.Fatalf("explicit empty validation_reasons should still serialize compactly, got %v", payload[1])
		}
	})
}

func TestWriteResultTextRendersCleanupRecommendations(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	result := contract.CleanupResult{
		DryRun: true,
		Filters: contract.CleanupFilters{
			State:     "stale",
			OlderThan: "72h",
			Project:   "cc9s",
		},
		Summary: contract.CleanupSummary{
			MatchedSessions: 3,
			MatchedProjects: 1,
			TotalSizeBytes:  4096,
			DeleteCount:     1,
			MaybeCount:      1,
			KeepCount:       1,
		},
		Projects: []contract.CleanupProjectGroup{{
			Name:         "cc9s",
			SessionCount: 3,
		}},
		Sessions: []contract.CleanupSessionMatch{
			{ID: "session-delete", Project: "cc9s", State: "Completed", Recommendation: "Delete", Reasons: []string{"small session file"}},
			{ID: "session-maybe", Project: "cc9s", State: "Completed", Recommendation: "Maybe", Reasons: []string{"moderate content"}},
			{ID: "session-keep", Project: "cc9s", State: "Idle", Recommendation: "Keep", Reasons: []string{"session is Idle"}},
		},
	}

	if err := WriteResult(&stdout, &stderr, contract.OutputText, result); err != nil {
		t.Fatalf("WriteResult returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty in text mode, got %q", stderr.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Recommendations",
		"Delete:   1 sessions",
		"Review:   1 sessions",
		"Keep:     1 sessions",
		"Tip",
		"\"--state completed\"",
		"Recommended for deletion (1)",
		"Review before deleting (1)",
		"Recommended to keep (1)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("cleanup text missing %q:\n%s", want, out)
		}
	}
}

func TestWriteResultTextRendersStatusHealthSummary(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	result := contract.StatusResult{
		Projects:  2,
		Sessions:  8,
		Resources: 3,
		Lifecycle: contract.LifecycleSummary{
			Active: 2,
		},
		Health: &claudefs.HealthMetrics{
			EnvironmentScore: 61,
			ProjectScores: []claudefs.ProjectHealthScore{
				{ProjectName: "helm-charts", HealthScore: 48, StaleRatio: 0.56, ActivityScore: 24},
			},
			Recommendations: []string{"Project helm-charts: cleanup recommended"},
		},
	}

	if err := WriteResult(&stdout, &stderr, contract.OutputText, result); err != nil {
		t.Fatalf("WriteResult returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty in text mode, got %q", stderr.String())
	}

	out := stdout.String()
	for _, unwanted := range []string{"Usage Trend", "Activity Overview", "Resource Usage", "Health Scores"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("status text should not include unexpected section %q:\n%s", unwanted, out)
		}
	}
	for _, want := range []string{"Environment Overview", "Health: 61/100", "Lowest Health Projects", "Recommendations"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status text missing %q:\n%s", want, out)
		}
	}
}
