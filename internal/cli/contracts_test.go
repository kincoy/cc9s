package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func TestWriteErrorJSONUsesStdout(t *testing.T) {
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	defer stdoutR.Close()

	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	defer stderrR.Close()

	writeError(stdoutW, stderrW, CLIError{Message: "boom"}, true)

	_ = stdoutW.Close()
	_ = stderrW.Close()

	stdoutData, err := io.ReadAll(stdoutR)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	stderrData, err := io.ReadAll(stderrR)
	if err != nil {
		t.Fatalf("read stderr: %v", err)
	}

	if got := strings.TrimSpace(string(stdoutData)); got != `{"error":"boom"}` {
		t.Fatalf("stdout = %q, want JSON error payload", got)
	}
	if len(stderrData) != 0 {
		t.Fatalf("stderr should be empty in json mode, got %q", string(stderrData))
	}
}

func TestRenderProjectListTextIncludesContractFields(t *testing.T) {
	var buf bytes.Buffer
	renderProjectListText(&buf, ProjectListResult{
		Projects: []ProjectListEntry{{
			Name:               "cc9s",
			SessionCount:       3,
			ActiveSessionCount: 1,
			SkillCount:         2,
			CommandCount:       1,
			AgentCount:         4,
			TotalSizeBytes:     2048,
			LastActiveAt:       "2026-03-25T10:00:00Z",
			Path:               "/tmp/cc9s",
		}},
	})

	out := buf.String()
	for _, want := range []string{"SKILLS", "COMMANDS", "AGENTS", "PATH", "/tmp/cc9s"} {
		if !strings.Contains(out, want) {
			t.Fatalf("project list output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderSkillAndAgentListTextIncludePaths(t *testing.T) {
	t.Run("skills", func(t *testing.T) {
		var buf bytes.Buffer
		renderSkillListText(&buf, SkillListResult{
			Skills: []SkillListEntry{{
				Name:    "openai-docs",
				Type:    "Skill",
				Scope:   "User",
				Status:  "Ready",
				Project: "",
				Path:    "/tmp/skills/openai-docs/SKILL.md",
			}},
		})

		out := buf.String()
		if !strings.Contains(out, "PATH") || !strings.Contains(out, "/tmp/skills/openai-docs/SKILL.md") {
			t.Fatalf("skill list output missing path field:\n%s", out)
		}
	})

	t.Run("agents", func(t *testing.T) {
		var buf bytes.Buffer
		renderAgentListText(&buf, AgentListResult{
			Agents: []AgentListEntry{{
				Name:    "reviewer",
				Scope:   "Project",
				Status:  "Ready",
				Project: "cc9s",
				Path:    "/tmp/agents/reviewer.md",
			}},
		})

		out := buf.String()
		if !strings.Contains(out, "PATH") || !strings.Contains(out, "/tmp/agents/reviewer.md") {
			t.Fatalf("agent list output missing path field:\n%s", out)
		}
	})
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

func TestRenderCleanupTextShowsRecommendationGroups(t *testing.T) {
	var buf bytes.Buffer
	renderCleanupText(&buf, CleanupResult{
		DryRun: true,
		Filters: CleanupFilters{
			State:     "stale",
			OlderThan: "72h",
			Project:   "cc9s",
		},
		Summary: CleanupSummary{
			MatchedSessions: 3,
			MatchedProjects: 1,
			TotalSizeBytes:  4096,
			DeleteCount:     1,
			MaybeCount:      1,
			KeepCount:       1,
		},
		Projects: []CleanupProjectGroup{{
			Name:         "cc9s",
			SessionCount: 3,
		}},
		Sessions: []CleanupSessionMatch{
			{
				ID:             "session-delete",
				Project:        "cc9s",
				State:          "Completed",
				Recommendation: "Delete",
				Reasons:        []string{"small session file"},
			},
			{
				ID:             "session-maybe",
				Project:        "cc9s",
				State:          "Completed",
				Recommendation: "Maybe",
				Reasons:        []string{"moderate content"},
			},
			{
				ID:             "session-keep",
				Project:        "cc9s",
				State:          "Idle",
				Recommendation: "Keep",
				Reasons:        []string{"session is Idle"},
			},
		},
	})

	out := buf.String()
	for _, want := range []string{
		"Recommendations",
		"Delete:   1 sessions",
		"Review:   1 sessions",
		"Keep:     1 sessions",
		"Tip",
		"usually only shows Delete",
		"\"--state completed\"",
		"Recommended for deletion (1)",
		"Review before deleting (1)",
		"Recommended to keep (1)",
		"small session file",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("cleanup text missing %q:\n%s", want, out)
		}
	}
}
