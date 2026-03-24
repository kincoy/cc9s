package ui

import (
	"testing"

	"github.com/kincoy/cc9s/internal/claudefs"
)

func TestAgentEditTargetPath(t *testing.T) {
	agent := claudefs.AgentResource{Name: "reviewer", Path: "/tmp/reviewer.md"}
	target, err := agentEditTargetPath(agent)
	if err != nil {
		t.Fatalf("agentEditTargetPath error: %v", err)
	}
	if target != "/tmp/reviewer.md" {
		t.Fatalf("target = %q, want /tmp/reviewer.md", target)
	}
}

func TestBuildAgentEditorCommandUsesProjectWorkingDir(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "nvim -u NONE")

	agent := claudefs.AgentResource{
		Name:        "reviewer",
		Path:        "/tmp/project/.claude/agents/reviewer.md",
		ProjectPath: "/tmp/project",
	}

	cmd, err := buildAgentEditorCommand(agent, agent.Path)
	if err != nil {
		t.Fatalf("buildAgentEditorCommand error: %v", err)
	}
	if got := cmd.Args[0]; got != "nvim" {
		t.Fatalf("editor command = %q, want nvim", got)
	}
	if got := cmd.Dir; got != "/tmp/project/.claude/agents" {
		t.Fatalf("working dir = %q, want /tmp/project/.claude/agents", got)
	}
}
