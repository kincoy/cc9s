package ui

import "testing"

func TestFooterHintsUseResourceCapabilities(t *testing.T) {
	registry := newResourceRegistry()

	projectHints := hintsForContext(FooterContext{
		Resource:   ResourceProjects,
		Descriptor: registry.MustGet(ResourceProjects),
	})
	if hasHint(projectHints, "0") {
		t.Fatal("projects footer should not advertise all-context shortcut")
	}

	sessionHints := hintsForContext(FooterContext{
		Resource:   ResourceSessions,
		Descriptor: registry.MustGet(ResourceSessions),
	})
	if !hasHint(sessionHints, "0") {
		t.Fatal("sessions footer should advertise all-context shortcut")
	}
	if !hasHint(sessionHints, "l") {
		t.Fatal("sessions footer should advertise logs support")
	}
}

func TestDetailOverlayHintsExposeEditOnlyForEditableResources(t *testing.T) {
	registry := newResourceRegistry()

	skillHints := hintsForContext(FooterContext{
		Resource:   ResourceSkills,
		Descriptor: registry.MustGet(ResourceSkills),
		Overlay:    OverlayDetail,
	})
	if !hasHint(skillHints, "e") {
		t.Fatal("skill detail overlay should advertise edit")
	}

	sessionHints := hintsForContext(FooterContext{
		Resource:   ResourceSessions,
		Descriptor: registry.MustGet(ResourceSessions),
		Overlay:    OverlayDetail,
	})
	if hasHint(sessionHints, "e") {
		t.Fatal("session detail overlay should not advertise edit")
	}
}

func hasHint(hints []KeyHint, key string) bool {
	for _, hint := range hints {
		if hint.Key == key {
			return true
		}
	}
	return false
}
