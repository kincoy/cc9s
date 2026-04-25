package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/kincoy/cc9s/internal/claudefs"
	cliroot "github.com/kincoy/cc9s/internal/cli/root"
	"github.com/kincoy/cc9s/internal/ui"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// extractGlobalFlags extracts global flags (--theme, --claude-dir) from args.
// Returns the extracted values and remaining args with flags removed.
func extractGlobalFlags(args []string) (theme, claudeDir string, rest []string) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--theme":
			if i+1 < len(args) {
				theme = args[i+1]
				i++ // skip value
			}
			continue
		case "--claude-dir":
			if i+1 < len(args) {
				claudeDir = args[i+1]
				i++ // skip value
			}
			continue
		}
		rest = append(rest, args[i])
	}

	// Theme fallback: CC9S_THEME env > "default"
	if theme == "" {
		if env := os.Getenv("CC9S_THEME"); env != "" {
			theme = env
		} else {
			theme = "default"
		}
	}

	// Claude dir fallback: CC9S_CLAUDE_DIR env is handled inside claudefs.ClaudeDir()
	// so we only pass the flag value if explicitly set.

	return theme, claudeDir, rest
}

func main() {
	themeName, claudeDir, args := extractGlobalFlags(os.Args[1:])
	styles.SetTheme(themeName)
	if claudeDir != "" {
		claudefs.SetClaudeDir(claudeDir)
	}

	if len(args) > 0 {
		os.Exit(cliroot.Execute(args))
	}

	p := tea.NewProgram(ui.NewAppModel())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
