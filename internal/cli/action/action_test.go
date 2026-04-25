package action

import (
	"errors"
	"testing"
	"time"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/cli/contract"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

func TestVersion(t *testing.T) {
	if got := Version(); got.Version == "" {
		t.Fatalf("version should not be empty")
	}
}

func TestThemesUsesCurrentTheme(t *testing.T) {
	got := Themes()
	if got.Current == "" {
		t.Fatalf("current theme should not be empty")
	}
	if len(got.Themes) == 0 {
		t.Fatalf("themes list should not be empty")
	}

	foundCurrent := false
	for _, theme := range got.Themes {
		if theme.Name == got.Current {
			foundCurrent = true
			if !theme.Current {
				t.Fatalf("theme %q should be marked current", theme.Name)
			}
		}
	}
	if !foundCurrent {
		t.Fatalf("current theme %q not found in list", got.Current)
	}
}

func TestProjectListUsesContractOptions(t *testing.T) {
	oldScanProjects := scanProjects
	oldFormatTime := formatTimeRFC3339
	t.Cleanup(func() {
		scanProjects = oldScanProjects
		formatTimeRFC3339 = oldFormatTime
	})

	scanProjects = func() claudefs.ScanResult {
		return claudefs.ScanResult{
			Projects: []claudefs.Project{{
				Name:               "alpha",
				SessionCount:       3,
				ActiveSessionCount: 1,
				LastActiveAt:       time.Date(2026, time.April, 1, 10, 0, 0, 0, time.UTC),
				SkillCount:         2,
				CommandCount:       1,
				AgentCount:         4,
				TotalSize:          2048,
				Path:               "/tmp/alpha",
			}},
		}
	}
	formatTimeRFC3339 = func(t time.Time) string {
		return t.Format(time.RFC3339)
	}

	got, err := ProjectsList(contract.ProjectListOptions{})
	if err != nil {
		t.Fatalf("ProjectsList returned error: %v", err)
	}
	if len(got.Projects) != 1 {
		t.Fatalf("projects length = %d, want 1", len(got.Projects))
	}
	if got := got.Projects[0].LastActiveAt; got != "2026-04-01T10:00:00Z" {
		t.Fatalf("last active = %q, want RFC3339", got)
	}
}

func TestSessionsCleanupDefaultsState(t *testing.T) {
	oldScanProjects := scanProjects
	oldLoadProjectSessions := loadProjectSessions
	oldNow := nowFunc
	oldAssess := quickAssessSession
	oldParseOlderThan := parseOlderThanDuration
	oldFormatTime := formatTimeRFC3339
	t.Cleanup(func() {
		scanProjects = oldScanProjects
		loadProjectSessions = oldLoadProjectSessions
		nowFunc = oldNow
		quickAssessSession = oldAssess
		parseOlderThanDuration = oldParseOlderThan
		formatTimeRFC3339 = oldFormatTime
	})

	scanProjects = func() claudefs.ScanResult {
		return claudefs.ScanResult{
			Projects: []claudefs.Project{{
				Name:         "alpha",
				EncodedPath:  "encoded-alpha",
				SessionCount: 1,
				TotalSize:    100,
			}},
		}
	}
	loadProjectSessions = func(string) ([]claudefs.Session, error) {
		return []claudefs.Session{{
			ID:          "session-1",
			ProjectPath: "/tmp/alpha",
			Lifecycle: claudefs.SessionLifecycleSnapshot{
				State: claudefs.SessionLifecycleStale,
			},
			LastActiveAt: time.Date(2026, time.March, 30, 10, 0, 0, 0, time.UTC),
			FileSize:     512,
		}}, nil
	}
	nowFunc = func() time.Time {
		return time.Date(2026, time.April, 1, 12, 0, 0, 0, time.UTC)
	}
	quickAssessSession = func(session claudefs.Session) claudefs.SessionAssessment {
		return claudefs.SessionAssessment{
			Recommendation: claudefs.RecommendDelete,
			Score:          7,
			Reasons:        []string{"stale"},
		}
	}
	parseOlderThanDuration = func(string) (time.Duration, error) {
		return 0, nil
	}
	formatTimeRFC3339 = func(t time.Time) string {
		return t.Format(time.RFC3339)
	}

	got, err := SessionsCleanup(contract.CleanupOptions{})
	if err != nil {
		t.Fatalf("SessionsCleanup returned error: %v", err)
	}
	if got.Filters.State != "stale" {
		t.Fatalf("default cleanup state = %q, want stale", got.Filters.State)
	}
	if len(got.Sessions) != 1 {
		t.Fatalf("cleanup sessions = %d, want 1", len(got.Sessions))
	}
}

func TestThemesCurrentNameMatchesStyles(t *testing.T) {
	got := Themes()
	if got.Current != styles.CurrentThemeName {
		t.Fatalf("current theme = %q, want %q", got.Current, styles.CurrentThemeName)
	}
}

func TestStatusSortsTopProjectsBySessionsThenActiveThenName(t *testing.T) {
	oldScanProjects := scanProjects
	oldLoadProjectSessions := loadProjectSessions
	oldComputeHealth := computeHealthMetrics
	oldScanSkills := scanSkills
	oldScanAgents := scanAgents
	t.Cleanup(func() {
		scanProjects = oldScanProjects
		loadProjectSessions = oldLoadProjectSessions
		computeHealthMetrics = oldComputeHealth
		scanSkills = oldScanSkills
		scanAgents = oldScanAgents
	})

	scanProjects = func() claudefs.ScanResult {
		return claudefs.ScanResult{
			Projects: []claudefs.Project{
				{Name: "gamma", EncodedPath: "gamma", SessionCount: 3, ActiveSessionCount: 1, TotalSize: 300},
				{Name: "alpha", EncodedPath: "alpha", SessionCount: 3, ActiveSessionCount: 2, TotalSize: 200},
				{Name: "beta", EncodedPath: "beta", SessionCount: 3, ActiveSessionCount: 2, TotalSize: 100},
			},
		}
	}
	loadProjectSessions = func(string) ([]claudefs.Session, error) { return nil, nil }
	computeHealthMetrics = func() (claudefs.HealthMetrics, error) {
		return claudefs.HealthMetrics{}, errors.New("skip health")
	}
	scanSkills = func() claudefs.SkillScanResult { return claudefs.SkillScanResult{} }
	scanAgents = func() claudefs.AgentScanResult { return claudefs.AgentScanResult{} }

	got, err := Status()
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if len(got.TopProjects) != 3 {
		t.Fatalf("top projects length = %d, want 3", len(got.TopProjects))
	}

	want := []string{"alpha", "beta", "gamma"}
	for i, name := range want {
		if got.TopProjects[i].Name != name {
			t.Fatalf("topProjects[%d] = %q, want %q", i, got.TopProjects[i].Name, name)
		}
	}
}

func TestParseOlderThanDurationSupportsDays(t *testing.T) {
	got, err := parseOlderThanDuration("7d")
	if err != nil {
		t.Fatalf("parseOlderThanDuration returned error: %v", err)
	}
	if got != 168*time.Hour {
		t.Fatalf("parseOlderThanDuration(7d) = %v, want %v", got, 168*time.Hour)
	}
}

func TestParseOlderThanDurationRejectsInvalidValues(t *testing.T) {
	for _, input := range []string{"abc", "-1d", "1.5d", "d"} {
		t.Run(input, func(t *testing.T) {
			if _, err := parseOlderThanDuration(input); err == nil {
				t.Fatalf("expected error for %q", input)
			}
		})
	}
}

func TestScopeMatches(t *testing.T) {
	tests := []struct {
		name  string
		value string
		query string
		want  bool
	}{
		{"exact match", "User", "user", true},
		{"exact match caps", "User", "USER", true},
		{"partial query in value", "Project", "proj", true},
		{"partial query in value", "Plugin", "lug", true},
		{"partial query in value", "Command", "mand", true},
		{"value in query", "Skill", "skills", true},
		{"no match", "User", "project", false},
		{"no match empty query", "User", "", false},
		{"no match empty value", "", "user", false},
		{"whitespace query trimmed", "User", "  user  ", true},
		{"single char match", "User", "u", true},
		{"single char no match", "User", "x", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scopeMatches(tt.value, tt.query); got != tt.want {
				t.Fatalf("scopeMatches(%q, %q) = %v, want %v", tt.value, tt.query, got, tt.want)
			}
		})
	}
}

func TestSortProjectEntries(t *testing.T) {
	entries := []contract.ProjectListEntry{
		{Name: "zebra", SessionCount: 1},
		{Name: "alpha", SessionCount: 5},
		{Name: "middle", SessionCount: 3},
	}

	t.Run("name", func(t *testing.T) {
		sorted := sortProjectEntries(entries, "name")
		if sorted[0].Name != "alpha" || sorted[1].Name != "middle" || sorted[2].Name != "zebra" {
			t.Fatalf("unexpected name sort order: %+v", sorted)
		}
		if entries[0].Name != "zebra" {
			t.Fatalf("original slice was modified")
		}
	})

	t.Run("sessions", func(t *testing.T) {
		sorted := sortProjectEntries(entries, "sessions")
		if sorted[0].SessionCount != 5 || sorted[1].SessionCount != 3 || sorted[2].SessionCount != 1 {
			t.Fatalf("unexpected session sort order: %+v", sorted)
		}
	})
}

func TestSortSessionEntriesUpdatedIsTimezoneSafe(t *testing.T) {
	entries := []contract.SessionListEntry{
		{LastActiveAt: "2026-03-25T11:00:00+08:00"},
		{LastActiveAt: "2026-03-25T04:00:00Z"},
		{LastActiveAt: ""},
	}

	sorted := sortSessionEntries(entries, "updated")
	if sorted[0].LastActiveAt != "2026-03-25T04:00:00Z" {
		t.Fatalf("first updated entry = %q, want latest UTC instant", sorted[0].LastActiveAt)
	}
	if sorted[2].LastActiveAt != "" {
		t.Fatalf("empty timestamp should sort last, got %q", sorted[2].LastActiveAt)
	}
}
