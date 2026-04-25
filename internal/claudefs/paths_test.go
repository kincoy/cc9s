package claudefs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeDir_Default(t *testing.T) {
	ResetClaudeDir()
	t.Cleanup(ResetClaudeDir)

	// Ensure no env override
	t.Setenv("CC9S_CLAUDE_DIR", "")

	dir := ClaudeDir()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(homeDir, ".claude")
	if dir != want {
		t.Errorf("ClaudeDir() = %q, want %q", dir, want)
	}
}

func TestClaudeDir_SetClaudeDir(t *testing.T) {
	ResetClaudeDir()
	t.Cleanup(ResetClaudeDir)

	SetClaudeDir("/tmp/custom-claude")
	dir := ClaudeDir()
	if dir != "/tmp/custom-claude" {
		t.Errorf("ClaudeDir() = %q, want %q", dir, "/tmp/custom-claude")
	}
}

func TestClaudeDir_EnvVar(t *testing.T) {
	ResetClaudeDir()
	t.Cleanup(ResetClaudeDir)

	t.Setenv("CC9S_CLAUDE_DIR", "/tmp/env-claude")
	dir := ClaudeDir()
	if dir != "/tmp/env-claude" {
		t.Errorf("ClaudeDir() = %q, want %q", dir, "/tmp/env-claude")
	}
}

func TestClaudeDir_SetOverridesEnv(t *testing.T) {
	ResetClaudeDir()
	t.Cleanup(ResetClaudeDir)

	t.Setenv("CC9S_CLAUDE_DIR", "/tmp/env-claude")
	SetClaudeDir("/tmp/flag-claude")
	dir := ClaudeDir()
	if dir != "/tmp/flag-claude" {
		t.Errorf("ClaudeDir() = %q, want %q (flag should override env)", dir, "/tmp/flag-claude")
	}
}

func TestSubDirs(t *testing.T) {
	ResetClaudeDir()
	t.Cleanup(ResetClaudeDir)

	SetClaudeDir("/tmp/test-claude")

	tests := []struct {
		name string
		fn   func() string
		want string
	}{
		{"ProjectsDir", ProjectsDir, "/tmp/test-claude/projects"},
		{"SessionsDir", SessionsDir, "/tmp/test-claude/sessions"},
		{"SkillsDir", SkillsDir, "/tmp/test-claude/skills"},
		{"CommandsDir", CommandsDir, "/tmp/test-claude/commands"},
		{"AgentsDir", AgentsDir, "/tmp/test-claude/agents"},
		{"PluginsDir", PluginsDir, "/tmp/test-claude/plugins"},
		{"PluginsInstalledPath", PluginsInstalledPath, "/tmp/test-claude/plugins/installed_plugins.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if got != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
