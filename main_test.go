package main

import (
	"testing"
)

func TestExtractThemeFlag(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantTheme string
		wantRest  []string
	}{
		{
			name:      "no args",
			args:      []string{},
			wantTheme: "default",
			wantRest:  nil,
		},
		{
			name:      "theme flag only",
			args:      []string{"--theme", "nord"},
			wantTheme: "nord",
			wantRest:  nil,
		},
		{
			name:      "theme with CLI args",
			args:      []string{"--theme", "solarized", "status"},
			wantTheme: "solarized",
			wantRest:  []string{"status"},
		},
		{
			name:      "CLI args without theme",
			args:      []string{"sessions", "list", "--json"},
			wantTheme: "default",
			wantRest:  []string{"sessions", "list", "--json"},
		},
		{
			name:      "theme at end",
			args:      []string{"status", "--theme", "dracula"},
			wantTheme: "dracula",
			wantRest:  []string{"status"},
		},
		{
			name:      "theme missing value",
			args:      []string{"status", "--theme"},
			wantTheme: "default",
			wantRest:  []string{"status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTheme, gotRest := extractThemeFlag(tt.args)
			if gotTheme != tt.wantTheme {
				t.Errorf("theme = %q, want %q", gotTheme, tt.wantTheme)
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

func TestExtractThemeFlag_EnvVar(t *testing.T) {
	t.Setenv("CC9S_THEME", "nord")
	gotTheme, gotRest := extractThemeFlag([]string{"status"})
	if gotTheme != "nord" {
		t.Errorf("theme = %q, want %q", gotTheme, "nord")
	}
	if len(gotRest) != 1 || gotRest[0] != "status" {
		t.Errorf("rest = %v, want [status]", gotRest)
	}
}

func TestExtractThemeFlag_FlagOverridesEnv(t *testing.T) {
	t.Setenv("CC9S_THEME", "nord")
	gotTheme, gotRest := extractThemeFlag([]string{"--theme", "dracula", "status"})
	if gotTheme != "dracula" {
		t.Errorf("theme = %q, want %q (flag should override env)", gotTheme, "dracula")
	}
	if len(gotRest) != 1 || gotRest[0] != "status" {
		t.Errorf("rest = %v, want [status]", gotRest)
	}
}
