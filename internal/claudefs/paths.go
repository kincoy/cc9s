package claudefs

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	globalClaudeDir string
	claudeDirOnce   sync.Once
	resolvedDir     string
)

// SetClaudeDir sets the global Claude data directory.
// Must be called before any path accessor is used (typically from main.go).
func SetClaudeDir(dir string) {
	globalClaudeDir = dir
}

// ClaudeDir returns the resolved Claude data directory.
// Priority: SetClaudeDir() value > CC9S_CLAUDE_DIR env > ~/.claude
func ClaudeDir() string {
	claudeDirOnce.Do(func() {
		if globalClaudeDir != "" {
			resolvedDir = globalClaudeDir
			return
		}
		if env := os.Getenv("CC9S_CLAUDE_DIR"); env != "" {
			resolvedDir = env
			return
		}
		homeDir, err := os.UserHomeDir()
		if err != nil {
			resolvedDir = filepath.Join(".", ".claude")
			return
		}
		resolvedDir = filepath.Join(homeDir, ".claude")
	})
	return resolvedDir
}

// ProjectsDir returns the projects directory path.
func ProjectsDir() string {
	return filepath.Join(ClaudeDir(), "projects")
}

// SessionsDir returns the active session markers directory path.
func SessionsDir() string {
	return filepath.Join(ClaudeDir(), "sessions")
}

// SkillsDir returns the user-level skills directory path.
func SkillsDir() string {
	return filepath.Join(ClaudeDir(), "skills")
}

// CommandsDir returns the user-level commands directory path.
func CommandsDir() string {
	return filepath.Join(ClaudeDir(), "commands")
}

// AgentsDir returns the user-level agents directory path.
func AgentsDir() string {
	return filepath.Join(ClaudeDir(), "agents")
}

// PluginsDir returns the plugins directory path.
func PluginsDir() string {
	return filepath.Join(ClaudeDir(), "plugins")
}

// PluginsInstalledPath returns the installed plugins JSON file path.
func PluginsInstalledPath() string {
	return filepath.Join(PluginsDir(), "installed_plugins.json")
}

// ResetClaudeDir resets the resolved directory cache (for testing only).
func ResetClaudeDir() {
	claudeDirOnce = sync.Once{}
	resolvedDir = ""
	globalClaudeDir = ""
}
