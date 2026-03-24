package ui

// DefaultResourceHooks provides item-specific behavior to the shared default list helper.
type DefaultResourceHooks[T any] struct {
	CursorKey    func(T) string
	InContext    func(T, Context) bool
	MatchesQuery func(T, string) bool
}

// DefaultResourceListState stores shared list-page mechanics for default-style resources.
type DefaultResourceListState[T any] struct {
	Context            Context
	FilterQuery        string
	Cursor             int
	Loading            bool
	RestoreCursorKey   string
	RestoreCursorIndex int

	AllItems     []T
	ContextItems []T
	VisibleItems []T
}

func NewDefaultResourceListState[T any]() DefaultResourceListState[T] {
	return DefaultResourceListState[T]{
		Context: Context{Type: ContextAll},
		Loading: true,
	}
}

func (s *DefaultResourceListState[T]) SetContext(ctx Context, hooks DefaultResourceHooks[T]) {
	s.Context = ctx
	s.FilterQuery = ""
	s.rebuild(hooks)
}

func (s *DefaultResourceListState[T]) SetItems(items []T, hooks DefaultResourceHooks[T]) {
	s.Loading = false
	s.AllItems = append([]T(nil), items...)
	s.rebuild(hooks)
	s.RestoreCursorAfterReload(hooks)
}

func (s *DefaultResourceListState[T]) ApplyFilter(query string, hooks DefaultResourceHooks[T]) {
	s.FilterQuery = query
	s.applyFilter(hooks)
}

func (s *DefaultResourceListState[T]) CaptureCursorForReload(hooks DefaultResourceHooks[T]) {
	s.RestoreCursorKey = ""
	s.RestoreCursorIndex = s.Cursor
	if s.Cursor >= 0 && s.Cursor < len(s.VisibleItems) && hooks.CursorKey != nil {
		s.RestoreCursorKey = hooks.CursorKey(s.VisibleItems[s.Cursor])
	}
}

func (s *DefaultResourceListState[T]) RestoreCursorAfterReload(hooks DefaultResourceHooks[T]) {
	defer func() {
		s.RestoreCursorKey = ""
		s.RestoreCursorIndex = 0
	}()

	if len(s.VisibleItems) == 0 {
		s.Cursor = 0
		return
	}
	if s.RestoreCursorKey != "" && hooks.CursorKey != nil {
		for i, item := range s.VisibleItems {
			if hooks.CursorKey(item) == s.RestoreCursorKey {
				s.Cursor = i
				return
			}
		}
	}
	s.Cursor = s.RestoreCursorIndex
	s.ClampCursor()
}

func (s *DefaultResourceListState[T]) FilterStats() (filtered, total int) {
	return len(s.VisibleItems), len(s.ContextItems)
}

func (s *DefaultResourceListState[T]) HasActiveFilter() bool {
	return normalizeResourceSearchQuery(s.FilterQuery) != ""
}

func (s *DefaultResourceListState[T]) ClampCursor() {
	if len(s.VisibleItems) == 0 {
		s.Cursor = 0
		return
	}
	if s.Cursor >= len(s.VisibleItems) {
		s.Cursor = len(s.VisibleItems) - 1
	}
	if s.Cursor < 0 {
		s.Cursor = 0
	}
}

func (s *DefaultResourceListState[T]) rebuild(hooks DefaultResourceHooks[T]) {
	if s.Context.Type == ContextAll || hooks.InContext == nil {
		s.ContextItems = append([]T(nil), s.AllItems...)
	} else {
		filtered := make([]T, 0, len(s.AllItems))
		for _, item := range s.AllItems {
			if hooks.InContext(item, s.Context) {
				filtered = append(filtered, item)
			}
		}
		s.ContextItems = filtered
	}
	s.applyFilter(hooks)
}

func (s *DefaultResourceListState[T]) applyFilter(hooks DefaultResourceHooks[T]) {
	query := normalizeResourceSearchQuery(s.FilterQuery)
	if query == "" || hooks.MatchesQuery == nil {
		s.VisibleItems = append([]T(nil), s.ContextItems...)
		s.ClampCursor()
		return
	}

	filtered := make([]T, 0, len(s.ContextItems))
	for _, item := range s.ContextItems {
		if hooks.MatchesQuery(item, query) {
			filtered = append(filtered, item)
		}
	}
	s.VisibleItems = filtered
	s.ClampCursor()
}
