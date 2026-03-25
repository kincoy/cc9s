package cli

import (
	"strings"
	"testing"
	"time"
)

// --- Parse: --scope flag ---

func TestParseScopeFlag(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{"skills list --scope user", []string{"skills", "list", "--scope", "user"}, "user", false},
		{"skills list --scope User", []string{"skills", "list", "--scope", "User"}, "User", false},
		{"agents list --scope project", []string{"agents", "list", "--scope", "project"}, "project", false},
		{"skills list --scope plugin", []string{"skills", "list", "--scope", "plugin"}, "plugin", false},

		// --scope on unsupported commands
		{"sessions list --scope user", []string{"sessions", "list", "--scope", "user"}, "", true},
		{"projects list --scope user", []string{"projects", "list", "--scope", "user"}, "", true},

		// missing value
		{"skills list --scope (no value)", []string{"skills", "list", "--scope"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := Parse(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "scope") {
					t.Errorf("error should mention scope, got: %s", err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cmd.ScopeFilter != tt.want {
				t.Errorf("ScopeFilter = %q, want %q", cmd.ScopeFilter, tt.want)
			}
		})
	}
}

// --- Parse: --type flag ---

func TestParseTypeFlag(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{"skills list --type command", []string{"skills", "list", "--type", "command"}, "command", false},
		{"skills list --type Skill", []string{"skills", "list", "--type", "Skill"}, "Skill", false},

		// --type on unsupported commands
		{"agents list --type command", []string{"agents", "list", "--type", "command"}, "", true},
		{"sessions list --type skill", []string{"sessions", "list", "--type", "skill"}, "", true},
		{"projects list --type skill", []string{"projects", "list", "--type", "skill"}, "", true},

		// missing value
		{"skills list --type (no value)", []string{"skills", "list", "--type"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := Parse(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "type") {
					t.Errorf("error should mention type, got: %s", err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cmd.TypeFilter != tt.want {
				t.Errorf("TypeFilter = %q, want %q", cmd.TypeFilter, tt.want)
			}
		})
	}
}

// --- Parse: --sort flag ---

func TestParseSortFlag(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		// valid fields for projects
		{"projects list --sort name", []string{"projects", "list", "--sort", "name"}, "name", false},
		{"projects list --sort sessions", []string{"projects", "list", "--sort", "sessions"}, "sessions", false},

		// valid fields for sessions
		{"sessions list --sort updated", []string{"sessions", "list", "--sort", "updated"}, "updated", false},
		{"sessions list --sort state", []string{"sessions", "list", "--sort", "state"}, "state", false},
		{"sessions list --sort project", []string{"sessions", "list", "--sort", "project"}, "project", false},

		// invalid field
		{"projects list --sort invalid", []string{"projects", "list", "--sort", "invalid"}, "", true},
		{"sessions list --sort name", []string{"sessions", "list", "--sort", "name"}, "", true},

		// --sort on unsupported commands
		{"skills list --sort name", []string{"skills", "list", "--sort", "name"}, "", true},
		{"agents list --sort name", []string{"agents", "list", "--sort", "name"}, "", true},

		// missing value
		{"projects list --sort (no value)", []string{"projects", "list", "--sort"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := Parse(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "sort") {
					t.Errorf("error should mention sort, got: %s", err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cmd.Sort != tt.want {
				t.Errorf("Sort = %q, want %q", cmd.Sort, tt.want)
			}
		})
	}
}

// --- Parse: --limit validation (strconv.Atoi fix) ---

func TestParseLimitValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    int
		wantErr bool
	}{
		{"valid limit", []string{"sessions", "list", "--limit", "10"}, 10, false},
		{"zero limit", []string{"sessions", "list", "--limit", "0"}, 0, false},
		{"invalid limit abc", []string{"sessions", "list", "--limit", "abc"}, 0, true},
		{"invalid limit 3.5", []string{"sessions", "list", "--limit", "3.5"}, 0, true},
		{"negative limit", []string{"sessions", "list", "--limit", "-1"}, -1, false},
		{"missing value", []string{"sessions", "list", "--limit"}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := Parse(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cmd.Limit != tt.want {
				t.Errorf("Limit = %d, want %d", cmd.Limit, tt.want)
			}
		})
	}
}

func TestParseRejectsUnsupportedFlagsForVerb(t *testing.T) {
	tests := []struct {
		name string
		args []string
		flag string
	}{
		{"projects inspect rejects limit", []string{"projects", "inspect", "cc9s", "--limit", "1"}, "--limit"},
		{"projects list rejects project filter", []string{"projects", "list", "--project", "cc9s"}, "--project"},
		{"sessions inspect rejects state", []string{"sessions", "inspect", "abc123", "--state", "stale"}, "--state"},
		{"sessions list rejects older-than", []string{"sessions", "list", "--older-than", "7d"}, "--older-than"},
		{"skills list rejects limit", []string{"skills", "list", "--limit", "5"}, "--limit"},
		{"agents inspect rejects scope", []string{"agents", "inspect", "reviewer", "--scope", "project"}, "--scope"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.args)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.flag) {
				t.Fatalf("error %q should mention %s", err.Error(), tt.flag)
			}
		})
	}
}

// --- Parse: --older-than validation (parse-time check) ---

func TestParseOlderThanValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{"valid 72h", []string{"sessions", "cleanup", "--dry-run", "--older-than", "72h"}, "72h", false},
		{"valid 7d", []string{"sessions", "cleanup", "--dry-run", "--older-than", "7d"}, "7d", false},
		{"valid 30d", []string{"sessions", "cleanup", "--dry-run", "--older-than", "30d"}, "30d", false},
		{"valid 30m", []string{"sessions", "cleanup", "--dry-run", "--older-than", "30m"}, "30m", false},
		{"invalid abc", []string{"sessions", "cleanup", "--dry-run", "--older-than", "abc"}, "", true},
		{"invalid empty", []string{"sessions", "cleanup", "--dry-run", "--older-than", ""}, "", true},
		{"missing value", []string{"sessions", "cleanup", "--dry-run", "--older-than"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := Parse(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cmd.OlderThan != tt.want {
				t.Errorf("OlderThan = %q, want %q", cmd.OlderThan, tt.want)
			}
		})
	}
}

// --- parseOlderThanDuration ---

func TestParseOlderThanDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"hours", "72h", 72 * time.Hour, false},
		{"minutes", "30m", 30 * time.Minute, false},
		{"seconds", "3600s", 3600 * time.Second, false},
		{"1 day", "1d", 24 * time.Hour, false},
		{"7 days", "7d", 168 * time.Hour, false},
		{"30 days", "30d", 720 * time.Hour, false},
		{"0 days", "0d", 0, false},
		{"mixed hours", "48h30m", 48*time.Hour + 30*time.Minute, false},
		{"invalid day value", "abc", 0, true},
		{"negative day", "-1d", 0, true},
		{"float day", "1.5d", 0, true},
		{"just d", "d", 0, true},
		{"empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOlderThanDuration(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseOlderThanDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// --- Parse: combined flags ---

func TestParseCombinedFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		check   func(*testing.T, *Command)
		wantErr bool
	}{
		{
			"skills list --scope user --type command",
			[]string{"skills", "list", "--scope", "user", "--type", "command"},
			func(t *testing.T, cmd *Command) {
				if cmd.ScopeFilter != "user" {
					t.Errorf("ScopeFilter = %q, want %q", cmd.ScopeFilter, "user")
				}
				if cmd.TypeFilter != "command" {
					t.Errorf("TypeFilter = %q, want %q", cmd.TypeFilter, "command")
				}
			},
			false,
		},
		{
			"skills list --project cc9s --scope project --type skill --json",
			[]string{"skills", "list", "--project", "cc9s", "--scope", "project", "--type", "skill", "--json"},
			func(t *testing.T, cmd *Command) {
				if cmd.ProjectFilter != "cc9s" {
					t.Errorf("ProjectFilter = %q", cmd.ProjectFilter)
				}
				if cmd.ScopeFilter != "project" {
					t.Errorf("ScopeFilter = %q", cmd.ScopeFilter)
				}
				if cmd.TypeFilter != "skill" {
					t.Errorf("TypeFilter = %q", cmd.TypeFilter)
				}
				if cmd.Output != OutputJSON {
					t.Errorf("Output = %d, want OutputJSON", cmd.Output)
				}
			},
			false,
		},
		{
			"sessions list --project cc9s --state active --sort updated --limit 3 --json",
			[]string{"sessions", "list", "--project", "cc9s", "--state", "active", "--sort", "updated", "--limit", "3", "--json"},
			func(t *testing.T, cmd *Command) {
				if cmd.ProjectFilter != "cc9s" {
					t.Errorf("ProjectFilter = %q", cmd.ProjectFilter)
				}
				if cmd.StateFilter != "active" {
					t.Errorf("StateFilter = %q", cmd.StateFilter)
				}
				if cmd.Sort != "updated" {
					t.Errorf("Sort = %q", cmd.Sort)
				}
				if cmd.Limit != 3 {
					t.Errorf("Limit = %d", cmd.Limit)
				}
				if cmd.Output != OutputJSON {
					t.Errorf("Output = %d, want OutputJSON", cmd.Output)
				}
			},
			false,
		},
		{
			"projects list --sort name --limit 10",
			[]string{"projects", "list", "--sort", "name", "--limit", "10"},
			func(t *testing.T, cmd *Command) {
				if cmd.Sort != "name" {
					t.Errorf("Sort = %q", cmd.Sort)
				}
				if cmd.Limit != 10 {
					t.Errorf("Limit = %d", cmd.Limit)
				}
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := Parse(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, cmd)
			}
		})
	}
}
