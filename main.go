package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/kincoy/cc9s/internal/cli"
	"github.com/kincoy/cc9s/internal/ui"
)

func main() {
	if len(os.Args) > 1 {
		os.Exit(cli.Run(os.Args[1:]))
	}

	p := tea.NewProgram(ui.NewAppModel())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
