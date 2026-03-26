package styles

import (
	"fmt"
	"os"
	"sort"

	"charm.land/lipgloss/v2"
)

// ThemeVersion is the current theme schema version.
const ThemeVersion = 1

// Theme defines a complete color palette for cc9s.
type Theme struct {
	Version         int    `yaml:"version"`
	Name            string `yaml:"name"`
	Active          string `yaml:"active"`
	Idle            string `yaml:"idle"`
	Completed       string `yaml:"completed"`
	Stale           string `yaml:"stale"`
	Normal          string `yaml:"normal"`
	Dim             string `yaml:"dim"`
	Highlight       string `yaml:"highlight"`
	Title           string `yaml:"title"`
	Accent          string `yaml:"accent"`
	Border          string `yaml:"border"`
	BorderFocus     string `yaml:"border_focus"`
	SearchBorder    string `yaml:"search_border"`
	Success         string `yaml:"success"`
	Background      string `yaml:"background,omitempty"`
	ForceBackground bool   `yaml:"force_background"`
	TableHeaderBg   string `yaml:"table_header_bg"`
}

// builtinThemes holds all built-in theme definitions, keyed by name.
var builtinThemes map[string]Theme

// CurrentThemeName tracks the currently active theme.
var CurrentThemeName = "default"

func init() {
	builtinThemes = map[string]Theme{
		"default":       themeDefault,
		"dark-solid":    themeDarkSolid,
		"high-contrast": themeHighContrast,
		"gruvbox":       themeGruvbox,
	}
}

// SetTheme activates the named theme. If the name is unknown, it falls back
// to "default" and prints a warning to stderr.
func SetTheme(name string) {
	t, ok := builtinThemes[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "cc9s: unknown theme %q, falling back to default\n", name)
		t = builtinThemes["default"]
		name = "default"
	}
	CurrentThemeName = name
	applyTheme(t)
}

// GetThemeByName returns the theme with the given name and whether it was found.
func GetThemeByName(name string) (Theme, bool) {
	t, ok := builtinThemes[name]
	return t, ok
}

// AvailableThemes returns the sorted names of all built-in themes.
func AvailableThemes() []string {
	names := make([]string, 0, len(builtinThemes))
	for n := range builtinThemes {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// applyTheme sets all package-level color variables from the theme definition,
// then rebuilds every lipgloss.Style that depends on those colors.
func applyTheme(t Theme) {
	ColorActive = lipgloss.Color(t.Active)
	ColorIdle = lipgloss.Color(t.Idle)
	ColorCompleted = lipgloss.Color(t.Completed)
	ColorStale = lipgloss.Color(t.Stale)
	ColorNormal = lipgloss.Color(t.Normal)
	ColorDim = lipgloss.Color(t.Dim)
	ColorHighlight = lipgloss.Color(t.Highlight)
	ColorTitle = lipgloss.Color(t.Title)
	ColorPurple = lipgloss.Color(t.Accent)
	ColorBorder = lipgloss.Color(t.Border)
	ColorBorderFocus = lipgloss.Color(t.BorderFocus)
	ColorSearchBorder = lipgloss.Color(t.SearchBorder)
	ColorFlashSuccess = lipgloss.Color(t.Success)
	ColorTableHeaderBg = lipgloss.Color(t.TableHeaderBg)
	ColorCommandBorder = lipgloss.Color(t.BorderFocus)

	// Backward compatibility: Warning reuses Idle color, Error reuses Stale color.
	ColorWarning = lipgloss.Color(t.Idle)
	ColorError = lipgloss.Color(t.Stale)

	if t.Background != "" {
		ColorBackground = lipgloss.Color(t.Background)
	} else {
		ColorBackground = nil
	}
	ForceBackground = t.ForceBackground

	rebuildStyles()
}
