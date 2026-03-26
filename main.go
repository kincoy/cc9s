package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/kincoy/cc9s/internal/cli"
	"github.com/kincoy/cc9s/internal/ui"
	"github.com/kincoy/cc9s/internal/ui/styles"
)

// extractThemeFlag extracts --theme <name> from args.
// Priority: --theme flag > CC9S_THEME env > "default".
// Returns the theme name and remaining args with --theme removed.
func extractThemeFlag(args []string) (string, []string) {
	var rest []string
	themeName := ""

	for i := 0; i < len(args); i++ {
		if args[i] == "--theme" {
			if i+1 < len(args) {
				themeName = args[i+1]
				i++ // skip value
			}
			// --theme without value: skip flag, use fallback
			continue
		}
		rest = append(rest, args[i])
	}

	if themeName != "" {
		return themeName, rest
	}

	if env := os.Getenv("CC9S_THEME"); env != "" {
		return env, rest
	}

	return "default", rest
}

func main() {
	themeName, args := extractThemeFlag(os.Args[1:])
	styles.SetTheme(themeName)

	if len(args) > 0 {
		os.Exit(cli.Run(args))
	}

	p := tea.NewProgram(ui.NewAppModel())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
