package root

import (
	"io"
	"testing"
)

func TestNewRootCommandRegistersTopLevelCommands(t *testing.T) {
	cmd := New(io.Discard, io.Discard)

	for _, name := range []string{
		"status", "projects", "sessions", "skills", "agents", "themes", "version", "help",
	} {
		found, _, err := cmd.Find([]string{name})
		if err != nil {
			t.Fatalf("find %q: %v", name, err)
		}
		if found == nil || found.Name() != name {
			t.Fatalf("command %q not registered correctly", name)
		}
		if name == "help" && found.Short != "Help about any command" {
			t.Fatalf("help command short = %q, want %q", found.Short, "Help about any command")
		}
	}
}

func TestNewRootCommandPreservesResourceAliases(t *testing.T) {
	cmd := New(io.Discard, io.Discard)

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
			t.Fatalf("find alias %q: %v", alias, err)
		}
		if found == nil || found.Name() != want {
			t.Fatalf("alias %q resolved to %q, want %q", alias, found.Name(), want)
		}
	}
}
