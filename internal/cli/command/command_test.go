package command

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kincoy/cc9s/internal/cli/contract"
)

func TestNewRegistersTopLevelCommandsAndAliases(t *testing.T) {
	cmd := New(&bytes.Buffer{}, &bytes.Buffer{})

	for _, name := range []string{"status", "projects", "sessions", "skills", "agents", "themes", "version", "help"} {
		if _, _, err := cmd.Find([]string{name}); err != nil {
			t.Fatalf("expected top-level command %q: %v", name, err)
		}
	}

	for alias, want := range map[string]string{
		"project": "projects",
		"proj":    "projects",
		"session": "sessions",
		"ss":      "sessions",
		"skill":   "skills",
		"sk":      "skills",
		"agent":   "agents",
		"ag":      "agents",
	} {
		found, _, err := cmd.Find([]string{alias})
		if err != nil {
			t.Fatalf("expected alias %q: %v", alias, err)
		}
		if found.Name() != want {
			t.Fatalf("alias %q resolved to %q, want %q", alias, found.Name(), want)
		}
	}
}

func TestStatusJSONExecutionRendersActionResult(t *testing.T) {
	oldRunStatus := runStatus
	t.Cleanup(func() { runStatus = oldRunStatus })

	runStatus = func() (contract.StatusResult, error) {
		return contract.StatusResult{
			Projects: 1,
			Sessions: 2,
			Lifecycle: contract.LifecycleSummary{
				Active: 1,
			},
		}, nil
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := New(&stdout, &stderr)
	cmd.SetArgs([]string{"status", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty, got %q", stderr.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("stdout is not json: %v\n%s", err, stdout.String())
	}
	if got := payload["projects"]; got != float64(1) {
		t.Fatalf("projects = %v, want 1", got)
	}
}

func TestProjectsDefaultCommandUsesListFlags(t *testing.T) {
	oldRunProjectsList := runProjectsList
	t.Cleanup(func() { runProjectsList = oldRunProjectsList })

	var captured contract.ProjectListOptions
	runProjectsList = func(opts contract.ProjectListOptions) (contract.ProjectListResult, error) {
		captured = opts
		return contract.ProjectListResult{}, nil
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := New(&stdout, &stderr)
	cmd.SetArgs([]string{"projects", "--limit", "3", "--sort", "sessions"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if captured.Limit != 3 || captured.Sort != "sessions" {
		t.Fatalf("captured opts = %+v, want limit=3 sort=sessions", captured)
	}
}

func TestSessionsImplicitInspectUsesBareID(t *testing.T) {
	oldRunSessionInspect := runSessionInspect
	t.Cleanup(func() { runSessionInspect = oldRunSessionInspect })

	var captured contract.SessionInspectOptions
	runSessionInspect = func(opts contract.SessionInspectOptions) (contract.SessionDetailResult, error) {
		captured = opts
		return contract.SessionDetailResult{
			SessionDetail: contract.SessionDetail{
				ID: opts.ID,
			},
		}, nil
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := New(&stdout, &stderr)
	cmd.SetArgs([]string{"sessions", "session-123", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if captured.ID != "session-123" {
		t.Fatalf("captured id = %q, want session-123", captured.ID)
	}
	if !strings.Contains(stdout.String(), `"id":"session-123"`) {
		t.Fatalf("stdout missing session id:\n%s", stdout.String())
	}
}

func TestSessionsCleanupRequiresDryRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := New(&stdout, &stderr)
	cmd.SetArgs([]string{"sessions", "cleanup"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error when --dry-run is missing")
	}
	if !strings.Contains(err.Error(), "--dry-run is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSessionsImplicitInspectRejectsListFlags(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "project",
			args: []string{"sessions", "session-123", "--project", "cc9s"},
			want: "--project is not supported for sessions inspect",
		},
		{
			name: "state",
			args: []string{"sessions", "session-123", "--state", "active"},
			want: "--state is not supported for sessions inspect",
		},
		{
			name: "limit",
			args: []string{"sessions", "session-123", "--limit", "1"},
			want: "--limit is not supported for sessions inspect",
		},
		{
			name: "sort",
			args: []string{"sessions", "session-123", "--sort", "updated"},
			want: "--sort is not supported for sessions inspect",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := New(&bytes.Buffer{}, &bytes.Buffer{})
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected invalid implicit inspect flag error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestVersionJSONExecutionRendersStructuredPayload(t *testing.T) {
	oldRunVersion := runVersion
	t.Cleanup(func() { runVersion = oldRunVersion })

	runVersion = func() contract.VersionResult {
		return contract.VersionResult{Version: "9.9.9"}
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := New(&stdout, &stderr)
	cmd.SetArgs([]string{"version", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty, got %q", stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != `{"version":"9.9.9"}` {
		t.Fatalf("stdout = %q, want structured version payload", strings.TrimSpace(stdout.String()))
	}
}

func TestVersionFlagExecutionMatchesVersionCommand(t *testing.T) {
	oldRunVersion := runVersion
	t.Cleanup(func() { runVersion = oldRunVersion })

	runVersion = func() contract.VersionResult {
		return contract.VersionResult{Version: "9.9.9"}
	}

	t.Run("text", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd := New(&stdout, &stderr)
		cmd.SetArgs([]string{"--version"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}
		if stderr.Len() != 0 {
			t.Fatalf("stderr should be empty, got %q", stderr.String())
		}
		if strings.TrimSpace(stdout.String()) != "cc9s v9.9.9" {
			t.Fatalf("stdout = %q, want version text", strings.TrimSpace(stdout.String()))
		}
	})

	t.Run("json", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd := New(&stdout, &stderr)
		cmd.SetArgs([]string{"--version", "--json"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}
		if stderr.Len() != 0 {
			t.Fatalf("stderr should be empty, got %q", stderr.String())
		}
		if strings.TrimSpace(stdout.String()) != `{"version":"9.9.9"}` {
			t.Fatalf("stdout = %q, want JSON version payload", strings.TrimSpace(stdout.String()))
		}
	})
}

func TestRootCommandRequiresSubcommandWhenNotAskingForVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := New(&stdout, &stderr)
	cmd.SetArgs([]string{"--json"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected missing command error")
	}
	if !strings.Contains(err.Error(), "expected a command") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectsListRejectsInvalidSort(t *testing.T) {
	cmd := New(&bytes.Buffer{}, &bytes.Buffer{})
	cmd.SetArgs([]string{"projects", "list", "--sort", "invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected invalid sort error")
	}
	if !strings.Contains(err.Error(), "invalid --sort field") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSessionsListRejectsInvalidSort(t *testing.T) {
	cmd := New(&bytes.Buffer{}, &bytes.Buffer{})
	cmd.SetArgs([]string{"sessions", "list", "--sort", "name"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected invalid sort error")
	}
	if !strings.Contains(err.Error(), "invalid --sort field") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSkillsListRejectsInvalidScopeAndTypeValues(t *testing.T) {
	t.Run("scope", func(t *testing.T) {
		cmd := New(&bytes.Buffer{}, &bytes.Buffer{})
		cmd.SetArgs([]string{"skills", "list", "--scope", "workspace"})

		err := cmd.Execute()
		if err == nil {
			t.Fatalf("expected invalid scope error")
		}
		if !strings.Contains(err.Error(), "invalid --scope value") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("type", func(t *testing.T) {
		cmd := New(&bytes.Buffer{}, &bytes.Buffer{})
		cmd.SetArgs([]string{"skills", "list", "--type", "agent"})

		err := cmd.Execute()
		if err == nil {
			t.Fatalf("expected invalid type error")
		}
		if !strings.Contains(err.Error(), "invalid --type value") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestCombinedFlagsFlowIntoActions(t *testing.T) {
	oldRunSkillsList := runSkillsList
	oldRunSessionsList := runSessionsList
	t.Cleanup(func() {
		runSkillsList = oldRunSkillsList
		runSessionsList = oldRunSessionsList
	})

	t.Run("skills list", func(t *testing.T) {
		var captured contract.SkillListOptions
		runSkillsList = func(opts contract.SkillListOptions) (contract.SkillListResult, error) {
			captured = opts
			return contract.SkillListResult{}, nil
		}

		cmd := New(&bytes.Buffer{}, &bytes.Buffer{})
		cmd.SetArgs([]string{"skills", "list", "--project", "cc9s", "--scope", "project", "--type", "skill", "--json"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}
		if captured.Project != "cc9s" || captured.Scope != "project" || captured.Type != "skill" {
			t.Fatalf("captured skill opts = %+v", captured)
		}
	})

	t.Run("sessions list", func(t *testing.T) {
		var captured contract.SessionListOptions
		runSessionsList = func(opts contract.SessionListOptions) (contract.SessionListResult, error) {
			captured = opts
			return contract.SessionListResult{}, nil
		}

		cmd := New(&bytes.Buffer{}, &bytes.Buffer{})
		cmd.SetArgs([]string{"sessions", "list", "--project", "cc9s", "--state", "active", "--sort", "updated", "--limit", "3", "--json"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}
		if captured.Project != "cc9s" || captured.State != "active" || captured.Sort != "updated" || captured.Limit != 3 {
			t.Fatalf("captured session opts = %+v", captured)
		}
	})
}
