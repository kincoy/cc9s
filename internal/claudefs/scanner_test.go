package claudefs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSummarizeProjectResources(t *testing.T) {
	projectRoot := t.TempDir()

	if err := os.MkdirAll(filepath.Join(projectRoot, ".claude", "skills", "release-check"), 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, ".claude", "skills", "review.md"), []byte("# Review"), 0o644); err != nil {
		t.Fatalf("write single-file skill: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectRoot, ".claude", "commands"), 0o755); err != nil {
		t.Fatalf("mkdir commands: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, ".claude", "commands", "deploy.md"), []byte("# Deploy"), 0o644); err != nil {
		t.Fatalf("write command: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectRoot, ".claude", "agents"), 0o755); err != nil {
		t.Fatalf("mkdir agents: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, ".claude", "agents", "reviewer.md"), []byte("# Reviewer"), 0o644); err != nil {
		t.Fatalf("write agent: %v", err)
	}

	summary := summarizeProjectResources(projectRoot)
	if summary.SkillCount != 2 {
		t.Fatalf("skill count = %d, want 2", summary.SkillCount)
	}
	if summary.CommandCount != 1 {
		t.Fatalf("command count = %d, want 1", summary.CommandCount)
	}
	if summary.AgentCount != 1 {
		t.Fatalf("agent count = %d, want 1", summary.AgentCount)
	}
	if !summary.HasSkillsRoot || !summary.HasCommandsRoot || !summary.HasAgentsRoot {
		t.Fatalf("expected all roots to exist, got %+v", summary)
	}
}
