package cli

import (
	"testing"
)

// --- scopeMatches helper ---

func TestScopeMatches(t *testing.T) {
	tests := []struct {
		name  string
		value string
		query string
		want  bool
	}{
		// exact match
		{"exact match", "User", "user", true},
		{"exact match caps", "User", "USER", true},
		// partial match: query contained in value
		{"partial query in value", "Project", "proj", true},
		{"partial query in value", "Plugin", "lug", true},
		{"partial query in value", "Command", "mand", true},
		// partial match: value contained in query
		{"value in query", "Skill", "skills", true},
		// no match
		{"no match", "User", "project", false},
		{"no match empty query", "User", "", false},
		{"no match empty value", "", "user", false},
		// edge cases
		{"whitespace query trimmed", "User", "  user  ", true},
		{"single char match", "User", "u", true},
		{"single char no match", "User", "x", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scopeMatches(tt.value, tt.query)
			if got != tt.want {
				t.Errorf("scopeMatches(%q, %q) = %v, want %v", tt.value, tt.query, got, tt.want)
			}
		})
	}
}

// --- sortProjectEntries ---

func TestSortProjectEntries(t *testing.T) {
	entries := []ProjectListEntry{
		{Name: "zebra", SessionCount: 1},
		{Name: "alpha", SessionCount: 5},
		{Name: "middle", SessionCount: 3},
	}

	t.Run("sort by name ascending", func(t *testing.T) {
		sorted := sortProjectEntries(entries, "name")
		if sorted[0].Name != "alpha" {
			t.Errorf("first entry = %q, want %q", sorted[0].Name, "alpha")
		}
		if sorted[1].Name != "middle" {
			t.Errorf("second entry = %q, want %q", sorted[1].Name, "middle")
		}
		if sorted[2].Name != "zebra" {
			t.Errorf("third entry = %q, want %q", sorted[2].Name, "zebra")
		}
		// original should not be modified
		if entries[0].Name != "zebra" {
			t.Error("original slice was modified")
		}
	})

	t.Run("sort by sessions descending", func(t *testing.T) {
		sorted := sortProjectEntries(entries, "sessions")
		if sorted[0].SessionCount != 5 {
			t.Errorf("first entry sessions = %d, want 5", sorted[0].SessionCount)
		}
		if sorted[1].SessionCount != 3 {
			t.Errorf("second entry sessions = %d, want 3", sorted[1].SessionCount)
		}
		if sorted[2].SessionCount != 1 {
			t.Errorf("third entry sessions = %d, want 1", sorted[2].SessionCount)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		sorted := sortProjectEntries(nil, "name")
		if len(sorted) != 0 {
			t.Errorf("expected empty, got %d entries", len(sorted))
		}
	})
}

func TestSortTopProjects(t *testing.T) {
	entries := []TopProject{
		{Name: "beta", Sessions: 3, Active: 1},
		{Name: "alpha", Sessions: 5, Active: 0},
		{Name: "gamma", Sessions: 5, Active: 2},
	}

	sortTopProjects(entries)

	if entries[0].Name != "gamma" {
		t.Fatalf("first entry = %q, want gamma", entries[0].Name)
	}
	if entries[1].Name != "alpha" {
		t.Fatalf("second entry = %q, want alpha", entries[1].Name)
	}
	if entries[2].Name != "beta" {
		t.Fatalf("third entry = %q, want beta", entries[2].Name)
	}
}

// --- sortSessionEntries ---

func TestSortSessionEntries(t *testing.T) {
	entries := []SessionListEntry{
		{Project: "zebra", State: "Idle", LastActiveAt: "2026-03-20T10:00:00+08:00"},
		{Project: "alpha", State: "Active", LastActiveAt: "2026-03-25T10:00:00+08:00"},
		{Project: "middle", State: "Stale", LastActiveAt: "2026-03-22T10:00:00+08:00"},
	}

	t.Run("sort by updated descending", func(t *testing.T) {
		sorted := sortSessionEntries(entries, "updated")
		if sorted[0].LastActiveAt != "2026-03-25T10:00:00+08:00" {
			t.Errorf("first = %q, want most recent", sorted[0].LastActiveAt)
		}
		if sorted[2].LastActiveAt != "2026-03-20T10:00:00+08:00" {
			t.Errorf("last = %q, want oldest", sorted[2].LastActiveAt)
		}
	})

	t.Run("sort by state ascending", func(t *testing.T) {
		sorted := sortSessionEntries(entries, "state")
		if sorted[0].State != "Active" {
			t.Errorf("first = %q, want Active", sorted[0].State)
		}
		if sorted[2].State != "Stale" {
			t.Errorf("last = %q, want Stale", sorted[2].State)
		}
	})

	t.Run("sort by project ascending", func(t *testing.T) {
		sorted := sortSessionEntries(entries, "project")
		if sorted[0].Project != "alpha" {
			t.Errorf("first = %q, want alpha", sorted[0].Project)
		}
		if sorted[2].Project != "zebra" {
			t.Errorf("last = %q, want zebra", sorted[2].Project)
		}
	})

	t.Run("empty timestamps sort last", func(t *testing.T) {
		withEmpty := []SessionListEntry{
			{LastActiveAt: "2026-03-20T10:00:00+08:00"},
			{LastActiveAt: ""},
			{LastActiveAt: "2026-03-25T10:00:00+08:00"},
		}
		sorted := sortSessionEntries(withEmpty, "updated")
		if sorted[0].LastActiveAt != "2026-03-25T10:00:00+08:00" {
			t.Errorf("first should be most recent, got %q", sorted[0].LastActiveAt)
		}
		if sorted[2].LastActiveAt != "" {
			t.Errorf("last should be empty, got %q", sorted[2].LastActiveAt)
		}
	})

	t.Run("all empty timestamps stable", func(t *testing.T) {
		allEmpty := []SessionListEntry{
			{LastActiveAt: ""},
			{LastActiveAt: ""},
		}
		sorted := sortSessionEntries(allEmpty, "updated")
		if len(sorted) != 2 {
			t.Errorf("expected 2 entries, got %d", len(sorted))
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		sorted := sortSessionEntries(nil, "updated")
		if len(sorted) != 0 {
			t.Errorf("expected empty, got %d entries", len(sorted))
		}
	})
}

// --- sortSessionEntries: timezone-safe comparison ---

func TestSortSessionEntriesTimezoneSafe(t *testing.T) {
	// C1 regression test: same instant with different timezone representations
	entries := []SessionListEntry{
		{LastActiveAt: "2026-03-25T11:00:00+08:00"}, // 03:00 UTC
		{LastActiveAt: "2026-03-25T04:00:00Z"},      // 04:00 UTC — should sort first (later)
	}

	sorted := sortSessionEntries(entries, "updated")
	if sorted[0].LastActiveAt != "2026-03-25T04:00:00Z" {
		t.Errorf("UTC time should sort first (later), but first = %q", sorted[0].LastActiveAt)
	}
}
