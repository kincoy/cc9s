package ui

import "testing"

type defaultListTestItem struct {
	key     string
	project string
	text    string
}

func testDefaultListHooks() DefaultResourceHooks[defaultListTestItem] {
	return DefaultResourceHooks[defaultListTestItem]{
		CursorKey: func(item defaultListTestItem) string { return item.key },
		InContext: func(item defaultListTestItem, ctx Context) bool {
			return item.project == "" || item.project == ctx.Value
		},
		MatchesQuery: func(item defaultListTestItem, query string) bool {
			return normalizeResourceSearchQuery(item.text) == query
		},
	}
}

func TestDefaultResourceListStateAppliesContextAndFilter(t *testing.T) {
	state := NewDefaultResourceListState[defaultListTestItem]()
	hooks := testDefaultListHooks()

	state.SetItems([]defaultListTestItem{
		{key: "global", project: "", text: "global"},
		{key: "alpha", project: "cc9s", text: "alpha"},
		{key: "beta", project: "other", text: "beta"},
	}, hooks)

	state.SetContext(Context{Type: ContextProject, Value: "cc9s"}, hooks)
	filtered, total := state.FilterStats()
	if filtered != 2 || total != 2 {
		t.Fatalf("stats = (%d, %d), want (2, 2)", filtered, total)
	}

	state.ApplyFilter("alpha", hooks)
	filtered, total = state.FilterStats()
	if filtered != 1 || total != 2 {
		t.Fatalf("stats after filter = (%d, %d), want (1, 2)", filtered, total)
	}
	if !state.HasActiveFilter() {
		t.Fatal("expected active filter after ApplyFilter")
	}
}

func TestDefaultResourceListStateRestoresCursorAfterReload(t *testing.T) {
	state := NewDefaultResourceListState[defaultListTestItem]()
	hooks := testDefaultListHooks()

	state.SetItems([]defaultListTestItem{
		{key: "alpha", text: "alpha"},
		{key: "beta", text: "beta"},
		{key: "gamma", text: "gamma"},
	}, hooks)
	state.Cursor = 1
	state.CaptureCursorForReload(hooks)

	state.SetItems([]defaultListTestItem{
		{key: "gamma", text: "gamma"},
		{key: "beta", text: "beta"},
		{key: "alpha", text: "alpha"},
	}, hooks)

	if state.Cursor != 1 {
		t.Fatalf("cursor = %d, want 1", state.Cursor)
	}
	if state.VisibleItems[state.Cursor].key != "beta" {
		t.Fatalf("selected key = %q, want beta", state.VisibleItems[state.Cursor].key)
	}
}
