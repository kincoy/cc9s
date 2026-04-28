package main

import (
	"testing"
)

func TestExtractGlobalFlags(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantTheme    string
		wantClaudeDir string
		wantRest     []string
	}{
		{
			name:         "no args",
			args:         []string{},
			wantTheme:    "default",
			wantClaudeDir: "",
			wantRest:     nil,
		},
		{
			name:         "theme flag only",
			args:         []string{"--theme", "nord"},
			wantTheme:    "nord",
			wantClaudeDir: "",
			wantRest:     nil,
		},
		{
			name:         "theme with CLI args",
			args:         []string{"--theme", "solarized", "status"},
			wantTheme:    "solarized",
			wantClaudeDir: "",
			wantRest:     []string{"status"},
		},
		{
			name:         "CLI args without theme",
			args:         []string{"sessions", "list", "--json"},
			wantTheme:    "default",
			wantClaudeDir: "",
			wantRest:     []string{"sessions", "list", "--json"},
		},
		{
			name:         "theme at end",
			args:         []string{"status", "--theme", "dracula"},
			wantTheme:    "dracula",
			wantClaudeDir: "",
			wantRest:     []string{"status"},
		},
		{
			name:         "theme missing value",
			args:         []string{"status", "--theme"},
			wantTheme:    "default",
			wantClaudeDir: "",
			wantRest:     []string{"status"},
		},
		{
			name:         "claude-dir flag",
			args:         []string{"--claude-dir", "/tmp/test", "status"},
			wantTheme:    "default",
			wantClaudeDir: "/tmp/test",
			wantRest:     []string{"status"},
		},
		{
			name:         "both flags",
			args:         []string{"--theme", "nord", "--claude-dir", "/tmp/test", "status"},
			wantTheme:    "nord",
			wantClaudeDir: "/tmp/test",
			wantRest:     []string{"status"},
		},
		{
			name:         "claude-dir missing value",
			args:         []string{"status", "--claude-dir"},
			wantTheme:    "default",
			wantClaudeDir: "",
			wantRest:     []string{"status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTheme, gotClaudeDir, gotRest := extractGlobalFlags(tt.args)
			if gotTheme != tt.wantTheme {
				t.Errorf("theme = %q, want %q", gotTheme, tt.wantTheme)
			}
			if gotClaudeDir != tt.wantClaudeDir {
				t.Errorf("claudeDir = %q, want %q", gotClaudeDir, tt.wantClaudeDir)
			}
			if len(gotRest) != len(tt.wantRest) {
				t.Errorf("rest = %v (len %d), want %v (len %d)", gotRest, len(gotRest), tt.wantRest, len(tt.wantRest))
				return
			}
			for i := range gotRest {
				if gotRest[i] != tt.wantRest[i] {
					t.Errorf("rest[%d] = %q, want %q", i, gotRest[i], tt.wantRest[i])
				}
			}
		})
	}
}

func TestExtractGlobalFlags_EnvVar(t *testing.T) {
	t.Setenv("CC9S_THEME", "nord")
	gotTheme, _, gotRest := extractGlobalFlags([]string{"status"})
	if gotTheme != "nord" {
		t.Errorf("theme = %q, want %q", gotTheme, "nord")
	}
	if len(gotRest) != 1 || gotRest[0] != "status" {
		t.Errorf("rest = %v, want [status]", gotRest)
	}
}

func TestExtractGlobalFlags_FlagOverridesEnv(t *testing.T) {
	t.Setenv("CC9S_THEME", "nord")
	gotTheme, _, gotRest := extractGlobalFlags([]string{"--theme", "dracula", "status"})
	if gotTheme != "dracula" {
		t.Errorf("theme = %q, want %q (flag should override env)", gotTheme, "dracula")
	}
	if len(gotRest) != 1 || gotRest[0] != "status" {
		t.Errorf("rest = %v, want [status]", gotRest)
	}
}
