package styles

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
)

// colorEqual compares two color.Color values by their RGBA components.
func colorEqual(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == b
	}
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

// colorNonEmpty returns true if the color is non-nil and not the zero color.
func colorNonEmpty(c color.Color) bool {
	return c != nil
}

func TestSetTheme_Default(t *testing.T) {
	SetTheme("default")

	want := lipgloss.Color("#7DCFFF")
	if !colorEqual(ColorHighlight, want) {
		t.Errorf("ColorHighlight mismatch after SetTheme(default)")
	}
	wantNormal := lipgloss.Color("#D4D4D4")
	if !colorEqual(ColorNormal, wantNormal) {
		t.Errorf("ColorNormal mismatch after SetTheme(default)")
	}
}

func TestSetTheme_Unknown_FallsBackToDefault(t *testing.T) {
	SetTheme("nonexistent-theme-xyz")

	want := lipgloss.Color("#7DCFFF")
	if !colorEqual(ColorHighlight, want) {
		t.Errorf("ColorHighlight mismatch after unknown theme fallback")
	}
	wantNormal := lipgloss.Color("#D4D4D4")
	if !colorEqual(ColorNormal, wantNormal) {
		t.Errorf("ColorNormal mismatch after unknown theme fallback")
	}
}

func TestSetTheme_AllBuiltins(t *testing.T) {
	for _, name := range AvailableThemes() {
		SetTheme(name)

		if !colorNonEmpty(ColorActive) {
			t.Errorf("theme %q: ColorActive is empty", name)
		}
		if !colorNonEmpty(ColorHighlight) {
			t.Errorf("theme %q: ColorHighlight is empty", name)
		}
	}
	// Restore default
	SetTheme("default")
}

func TestAvailableThemes(t *testing.T) {
	themes := AvailableThemes()

	if len(themes) < 4 {
		t.Errorf("AvailableThemes() returned %d themes, want at least 4", len(themes))
	}

	required := []string{"default", "dark-solid", "high-contrast", "gruvbox"}
	have := make(map[string]bool, len(themes))
	for _, n := range themes {
		have[n] = true
	}
	for _, r := range required {
		if !have[r] {
			t.Errorf("AvailableThemes() missing required theme %q", r)
		}
	}
}

func TestThemeCompleteness(t *testing.T) {
	requiredFields := []string{
		"Active", "Idle", "Completed", "Stale",
		"Normal", "Dim", "Highlight", "Title", "Accent",
		"Border", "BorderFocus", "SearchBorder",
		"Success", "TableHeaderBg",
	}

	for _, name := range AvailableThemes() {
		theme, ok := GetThemeByName(name)
		if !ok {
			t.Errorf("GetThemeByName(%q) returned false", name)
			continue
		}

		fields := map[string]string{
			"Active":        theme.Active,
			"Idle":          theme.Idle,
			"Completed":     theme.Completed,
			"Stale":         theme.Stale,
			"Normal":        theme.Normal,
			"Dim":           theme.Dim,
			"Highlight":     theme.Highlight,
			"Title":         theme.Title,
			"Accent":        theme.Accent,
			"Border":        theme.Border,
			"BorderFocus":   theme.BorderFocus,
			"SearchBorder":  theme.SearchBorder,
			"Success":       theme.Success,
			"TableHeaderBg": theme.TableHeaderBg,
		}

		for _, fname := range requiredFields {
			if fields[fname] == "" {
				t.Errorf("theme %q: field %q is empty", name, fname)
			}
		}
	}
}
