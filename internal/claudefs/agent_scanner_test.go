package claudefs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanAgentRootsParsesReadyAndInvalidCandidates(t *testing.T) {
	root := t.TempDir()

	readyContent := `---
name: reviewer
description: Reviews Go changes
tools:
  - Read
model: sonnet
---
Body text.
`
	invalidContent := "# Missing frontmatter\n\nStill visible.\n"

	if err := os.WriteFile(filepath.Join(root, "reviewer.md"), []byte(readyContent), 0o644); err != nil {
		t.Fatalf("write ready agent: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "broken.md"), []byte(invalidContent), 0o644); err != nil {
		t.Fatalf("write invalid agent: %v", err)
	}

	agents, err := scanAgentRoots([]AgentDiscoveryRoot{{Path: root, Source: AgentSourceUser}})
	if err != nil {
		t.Fatalf("scan agents: %v", err)
	}
	if len(agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(agents))
	}

	var foundReady, foundInvalid bool
	for _, agent := range agents {
		switch agent.Name {
		case "reviewer":
			foundReady = true
			if agent.Description != "Reviews Go changes" {
				t.Fatalf("description = %q", agent.Description)
			}
			if agent.Summary != "Reviews Go changes" {
				t.Fatalf("summary = %q", agent.Summary)
			}
			if agent.Model != "sonnet" {
				t.Fatalf("model = %q", agent.Model)
			}
			if len(agent.Tools) != 1 || agent.Tools[0] != "Read" {
				t.Fatalf("tools = %#v", agent.Tools)
			}
		case "broken":
			foundInvalid = true
			if !strings.Contains(strings.Join(agent.ValidationReasons, " "), "Missing required identifying metadata") {
				t.Fatalf("expected missing metadata reason, got %#v", agent.ValidationReasons)
			}
		}
	}

	if !foundReady || !foundInvalid {
		t.Fatalf("expected both ready and invalid candidates to be found")
	}
}

func TestUniqueAgentDiscoveryRootsDeduplicatesPaths(t *testing.T) {
	root := t.TempDir()
	roots := uniqueAgentDiscoveryRoots([]AgentDiscoveryRoot{
		{Path: root, Source: AgentSourceUser},
		{Path: filepath.Join(root, "."), Source: AgentSourceProject, ProjectName: "dup"},
	})

	if len(roots) != 1 {
		t.Fatalf("expected duplicate roots to collapse to 1, got %d", len(roots))
	}
	if roots[0].Source != AgentSourceUser {
		t.Fatalf("expected first root to win, got %s", roots[0].Source)
	}
}

func TestReconcileAgentRecognitionSupportsUserProjectAndPluginAgents(t *testing.T) {
	agents := []AgentResource{
		{Name: "user-agent", Source: AgentSourceUser, Path: "/tmp/user-agent.md", ValidationReasons: []string{"Summary derived from description metadata"}},
		{Name: "project-agent", Source: AgentSourceProject, ProjectPath: "/work/project-a", ProjectName: "project-a", Path: "/work/project-a/.claude/agents/project-agent.md", ValidationReasons: []string{"Summary derived from description metadata"}},
		{Name: "go-reviewer", Source: AgentSourcePlugin, PluginName: "everything-claude-code", Path: "/tmp/plugin/go-reviewer.md", ValidationReasons: []string{"Summary derived from description metadata"}},
		{Name: "local-plugin", Source: AgentSourcePlugin, PluginName: "interface-design", ProjectPath: "/work/project-a", ProjectName: "project-a", Path: "/tmp/plugin/local-plugin.md", ValidationReasons: []string{"Summary derived from description metadata"}},
		{Name: "broken-agent", Source: AgentSourceProject, ProjectPath: "/work/project-a", ProjectName: "project-a", Path: "/work/project-a/.claude/agents/broken-agent.md", ValidationReasons: []string{"Missing required identifying metadata"}},
	}

	origRunner := runClaudeAgentsCommand
	t.Cleanup(func() { runClaudeAgentsCommand = origRunner })
	runClaudeAgentsCommand = func(workDir string, settingSources string) (string, error) {
		switch {
		case workDir == "" && settingSources == "user":
			return `17 active agents

Custom agents:
  user-agent · sonnet

Plugin agents:
  everything-claude-code:go-reviewer · sonnet

Built-in agents:
  general-purpose · inherit
`, nil
		case workDir == "/work/project-a" && settingSources == "project,local":
			return `6 active agents

Custom agents:
  project-agent · sonnet

Plugin agents:
  interface-design:local-plugin · sonnet

Built-in agents:
  Explore · haiku
`, nil
		default:
			return "", fmt.Errorf("unexpected lookup: %q %q", workDir, settingSources)
		}
	}

	reconciled, err := reconcileAgentRecognition(agents)
	if err != nil {
		t.Fatalf("reconcile agents: %v", err)
	}

	statusByPath := make(map[string]AgentStatus, len(reconciled))
	reasonsByPath := make(map[string][]string, len(reconciled))
	for _, agent := range reconciled {
		statusByPath[agent.Path] = agent.Status
		reasonsByPath[agent.Path] = agent.ValidationReasons
	}

	if statusByPath["/tmp/user-agent.md"] != AgentStatusReady {
		t.Fatalf("user agent status = %s", statusByPath["/tmp/user-agent.md"])
	}
	if statusByPath["/work/project-a/.claude/agents/project-agent.md"] != AgentStatusReady {
		t.Fatalf("project agent status = %s", statusByPath["/work/project-a/.claude/agents/project-agent.md"])
	}
	if statusByPath["/tmp/plugin/go-reviewer.md"] != AgentStatusReady {
		t.Fatalf("plugin agent status = %s", statusByPath["/tmp/plugin/go-reviewer.md"])
	}
	if statusByPath["/tmp/plugin/local-plugin.md"] != AgentStatusReady {
		t.Fatalf("local plugin agent status = %s", statusByPath["/tmp/plugin/local-plugin.md"])
	}
	if statusByPath["/work/project-a/.claude/agents/broken-agent.md"] != AgentStatusInvalid {
		t.Fatalf("broken agent status = %s", statusByPath["/work/project-a/.claude/agents/broken-agent.md"])
	}
	if !strings.Contains(strings.Join(reasonsByPath["/work/project-a/.claude/agents/broken-agent.md"], " "), "Agent file is not recognized by Claude Code") {
		t.Fatalf("expected recognition failure reason, got %#v", reasonsByPath["/work/project-a/.claude/agents/broken-agent.md"])
	}
}

func TestReconcileAgentRecognitionReturnsLookupFailure(t *testing.T) {
	origRunner := runClaudeAgentsCommand
	t.Cleanup(func() { runClaudeAgentsCommand = origRunner })
	runClaudeAgentsCommand = func(workDir string, settingSources string) (string, error) {
		return "", fmt.Errorf("lookup failed")
	}

	_, err := reconcileAgentRecognition([]AgentResource{
		{Name: "user-agent", Source: AgentSourceUser, Path: "/tmp/user-agent.md"},
	})
	if err == nil {
		t.Fatal("expected lookup failure")
	}
}

func TestParseAgentRecognitionOutputIgnoresBuiltIns(t *testing.T) {
	snapshot, err := parseAgentRecognitionOutput(`4 active agents

Plugin agents:
  sample:planner · opus

Built-in agents:
  Plan · inherit
  Explore · haiku
`)
	if err != nil {
		t.Fatalf("parse recognition: %v", err)
	}

	if len(snapshot.PluginAgents) != 1 {
		t.Fatalf("expected 1 plugin agent, got %d", len(snapshot.PluginAgents))
	}
	if len(snapshot.CustomAgents) != 0 {
		t.Fatalf("expected 0 custom agents, got %d", len(snapshot.CustomAgents))
	}
}

func TestParseAgentRecognitionOutputRejectsUnexpectedContent(t *testing.T) {
	_, err := parseAgentRecognitionOutput(`warning: something changed
4 active agents

Built-in agents:
  Explore · haiku
`)
	if err == nil {
		t.Fatal("expected parse failure for unexpected content")
	}
}

func TestParseAgentRecognitionOutputAcceptsProjectAgentsSection(t *testing.T) {
	snapshot, err := parseAgentRecognitionOutput(`6 active agents

Project agents:
  reviewer · sonnet

Plugin agents:
  sample:planner · opus

Built-in agents:
  Explore · haiku
`)
	if err != nil {
		t.Fatalf("expected project agents section to parse, got %v", err)
	}
	if _, ok := snapshot.CustomAgents["reviewer"]; !ok {
		t.Fatal("expected reviewer in custom agents set")
	}
	if _, ok := snapshot.PluginAgents["sample:planner"]; !ok {
		t.Fatal("expected sample:planner in plugin agents set")
	}
}

func TestReconcileAgentRecognitionReturnsParseFailure(t *testing.T) {
	origRunner := runClaudeAgentsCommand
	t.Cleanup(func() { runClaudeAgentsCommand = origRunner })
	runClaudeAgentsCommand = func(workDir string, settingSources string) (string, error) {
		return `warning: unexpected format
4 active agents
`, nil
	}

	_, err := reconcileAgentRecognition([]AgentResource{
		{Name: "user-agent", Source: AgentSourceUser, Path: "/tmp/user-agent.md"},
	})
	if err == nil {
		t.Fatal("expected parse failure")
	}
}
