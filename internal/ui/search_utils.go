package ui

import "strings"

func normalizeResourceSearchQuery(query string) string {
	q := strings.TrimSpace(strings.ToLower(query))
	q = strings.TrimPrefix(q, "/")
	return strings.TrimSpace(q)
}
