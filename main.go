package main

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/kincoy/cc9s/internal/ui"
)

func main() {
	p := tea.NewProgram(ui.NewAppModel())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
