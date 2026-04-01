package action

import (
	"github.com/kincoy/cc9s/internal/cli/contract"
	"github.com/kincoy/cc9s/internal/ui/styles"
	"github.com/kincoy/cc9s/internal/version"
)

// Version returns the current release version.
func Version() contract.VersionResult {
	return contract.VersionResult{Version: version.Version}
}

// Themes returns the available themes and marks the current one.
func Themes() contract.ThemesResult {
	names := styles.AvailableThemes()
	descriptions := themeDescriptions()
	entries := make([]contract.ThemeEntry, 0, len(names))
	for _, name := range names {
		entries = append(entries, contract.ThemeEntry{
			Name:        name,
			Description: descriptions[name],
			Current:     name == styles.CurrentThemeName,
		})
	}
	return contract.ThemesResult{
		Themes:  entries,
		Current: styles.CurrentThemeName,
	}
}

func themeDescriptions() map[string]string {
	return map[string]string{
		"default":       "Optimized default, transparent-friendly",
		"dark-solid":    "Forced dark background",
		"high-contrast": "Maximum readability",
		"gruvbox":       "Warm retro palette",
	}
}
