package claudefs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var DefaultActivityWindow = ActivityWindow{
	ActiveWindow: 15 * time.Minute,
	IdleWindow:   24 * time.Hour,
}

// DecodePathFS decodes a Claude project path via filesystem validation.
//
// Claude Code path encoding rules: / -> -, . -> -, - -> - (unchanged)
// All three characters map to "-", so pure string decoding is not reversible.
//
// Algorithm: longest match + filesystem validation
// 1. Split the encoded string by "-"
// 2. Starting from root, try consuming 1-4 segments at a time
// 3. For each group of segments, generate all candidate directory names joined by "." or "-"
// 4. Use os.Stat to check which candidate directories exist
// 5. Prefer the longest match (to avoid short matches stealing from longer ones, e.g. kubernetes vs kubernetes-sigs)
//
// Limitation: when multiple valid paths coexist (e.g. /a/ and /a-b/ both exist),
// the wrong one may be chosen. Callers should fall back to JSONL cwd to correct this.
func DecodePathFS(encodedPath string) string {
	if encodedPath == "" {
		return ""
	}

	segments := strings.Split(encodedPath, "-")

	// Skip leading empty segment (encoded path starts with "-", representing root "/")
	if len(segments) > 0 && segments[0] == "" {
		segments = segments[1:]
	}

	if len(segments) == 0 {
		return "/"
	}

	path := "/"
	idx := 0

	for idx < len(segments) {
		maxTry := len(segments) - idx
		if maxTry > 4 {
			maxTry = 4
		}

		// Collect matches of all lengths, prefer the longest one
		bestPath := ""
		bestTryLen := 0

		for tryLen := 1; tryLen <= maxTry; tryLen++ {
			// Empty segment cannot be used as a directory name on its own ("." would match all directories)
			if segments[idx] == "" && tryLen == 1 {
				continue
			}

			group := segments[idx : idx+tryLen]
			candidates := buildCandidates(group)

			for _, candidate := range candidates {
				testPath := path + candidate
				info, err := os.Stat(testPath)
				if err == nil && info.IsDir() {
					if tryLen > bestTryLen {
						bestPath = testPath + "/"
						bestTryLen = tryLen
					}
					break // First match at this length is sufficient
				}
			}
		}

		if bestTryLen > 0 {
			path = bestPath
			idx += bestTryLen
		} else {
			// Fallback: use the current segment directly as a path component
			if segments[idx] == "" {
				// Empty segment represents ".", merge with the next segment
				if idx+1 < len(segments) {
					path += "." + segments[idx+1] + "/"
					idx += 2
				} else {
					idx++
				}
			} else {
				path += segments[idx] + "/"
				idx++
			}
		}
	}

	// Remove trailing "/", but preserve root "/"
	result := strings.TrimSuffix(path, "/")
	if result == "" {
		return "/"
	}
	return result
}

// buildCandidates generates all possible directory name candidates for a group of segments.
// Each pair of adjacent segments tries "." and "-" as separators, producing 2^(n-1) combinations.
func buildCandidates(segments []string) []string {
	n := len(segments)
	if n == 0 {
		return nil
	}
	if n == 1 {
		if segments[0] == "" {
			return nil
		}
		return []string{segments[0]}
	}

	numSeps := n - 1
	candidates := make([]string, 0, 1<<uint(numSeps))

	for mask := 0; mask < (1 << uint(numSeps)); mask++ {
		var b strings.Builder
		for i, s := range segments {
			if i > 0 {
				if mask&(1<<uint(i-1)) != 0 {
					b.WriteByte('.')
				} else {
					b.WriteByte('-')
				}
			}
			b.WriteString(s)
		}
		candidates = append(candidates, b.String())
	}

	return candidates
}

// ExtractProjectName extracts the project name from a path (last segment).
// Input:  "/Users/kinco/go/src/github.com/kincoy/cc9s"
// Output: "cc9s"
func ExtractProjectName(path string) string {
	return filepath.Base(path)
}

// FormatSize formats a byte size into a human-readable string.
// Examples: "1.2 MB", "340 KB", "52 B"
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatTimeAgo formats a time as a relative duration.
// Examples: "2h ago", "3d ago", "2 weeks ago", "2025-12-01"
func FormatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	now := time.Now()
	duration := now.Sub(t)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	case duration < 30*24*time.Hour:
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	default:
		// Over a month ago, show the actual date
		return t.Format("2006-01-02")
	}
}

// FormatSessionID formats a session ID for display.
// Input:  "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
// Output: "a1b2c3d4...890" (first 8 chars + ... + last 3 chars)
func FormatSessionID(id string, maxLen int) string {
	if len(id) <= maxLen {
		return id
	}

	// Default format: first 8 chars + ... + last 3 chars
	if len(id) > 11 {
		return id[:8] + "..." + id[len(id)-3:]
	}

	return id
}

// FormatEventCount formats an event count.
// Input: 342 -> "342"
// Input: 1234 -> "1.2k"
func FormatEventCount(count int) string {
	if count < 1000 {
		return fmt.Sprintf("%d", count)
	}

	if count < 10000 {
		return fmt.Sprintf("%.1fk", float64(count)/1000)
	}

	return fmt.Sprintf("%.0fk", float64(count)/1000)
}

// FormatTime formats an absolute time.
// Example: "2026-03-21 15:30:45"
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Format("2006-01-02 15:04:05")
}

// FormatDuration formats a duration.
// Examples: "2h 30m", "45m 30s", "30s"
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return "< 1s"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	if minutes > 0 {
		if seconds > 0 {
			return fmt.Sprintf("%dm %ds", minutes, seconds)
		}
		return fmt.Sprintf("%dm", minutes)
	}

	return fmt.Sprintf("%ds", seconds)
}

// FormatNumber formats a number with thousands separators.
// Input: 1234567 -> "1,234,567"
func FormatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	// Simple thousands separator formatting
	str := fmt.Sprintf("%d", n)
	var result strings.Builder
	length := len(str)

	for i, c := range str {
		if i > 0 && (length-i)%3 == 0 {
			result.WriteByte(',')
		}
		result.WriteRune(c)
	}

	return result.String()
}

// NormalizeActivityWindow fills unset lifecycle thresholds with defaults.
func NormalizeActivityWindow(window ActivityWindow) ActivityWindow {
	if window.ActiveWindow <= 0 {
		window.ActiveWindow = DefaultActivityWindow.ActiveWindow
	}
	if window.IdleWindow <= window.ActiveWindow {
		window.IdleWindow = DefaultActivityWindow.IdleWindow
	}
	return window
}

// ClassifySessionLifecycle derives the shared lifecycle snapshot for a session.
func ClassifySessionLifecycle(lastActiveAt time.Time, hasActiveMarker bool, health SessionHealth, now time.Time, window ActivityWindow) SessionLifecycleSnapshot {
	window = NormalizeActivityWindow(window)

	snapshot := SessionLifecycleSnapshot{
		ClassifiedAt: now,
	}

	evidence := StateEvidenceSummary{
		LastActiveAt:    lastActiveAt,
		HasActiveMarker: hasActiveMarker,
	}

	if !health.IsReliable {
		snapshot.State = SessionLifecycleStale
		evidence.State = snapshot.State
		evidence.Reasons = []string{
			"This session is still listed, but it no longer looks like a reliable normal session.",
			healthProblemReason(health.Problem),
			"Stale means abnormal residue, not just an old inactive session.",
		}
		snapshot.Evidence = evidence
		return snapshot
	}

	if lastActiveAt.IsZero() {
		snapshot.State = SessionLifecycleIdle
		evidence.State = snapshot.State
		evidence.Reasons = []string{
			"This session still looks structurally healthy.",
			"No trustworthy last-activity timestamp is available, so it is treated as not currently active.",
			"It remains a normal session rather than stale residue.",
		}
		snapshot.Evidence = evidence
		return snapshot
	}

	if now.Before(lastActiveAt) {
		now = lastActiveAt
		snapshot.ClassifiedAt = now
	}

	age := now.Sub(lastActiveAt)

	switch {
	case age <= window.ActiveWindow:
		snapshot.State = SessionLifecycleActive
		evidence.Reasons = []string{
			fmt.Sprintf("Last activity was %s, inside the %s active window.", FormatTimeAgo(lastActiveAt), FormatDuration(window.ActiveWindow)),
			activeCoherenceReason(hasActiveMarker),
		}
	case hasActiveMarker:
		snapshot.State = SessionLifecycleIdle
		evidence.Reasons = []string{
			"Recent activity is no longer current, so this session is not active right now.",
			"The session still has a live marker and remains structurally healthy.",
			"Long inactivity with a healthy session is treated as idle, not stale.",
		}
	case age <= window.IdleWindow:
		snapshot.State = SessionLifecycleIdle
		evidence.Reasons = []string{
			fmt.Sprintf("Last activity was %s, outside the %s active window.", FormatTimeAgo(lastActiveAt), FormatDuration(window.ActiveWindow)),
			fmt.Sprintf("It is still within the %s idle window.", FormatDuration(window.IdleWindow)),
			"This session remains healthy and can still be treated as normal.",
		}
	default:
		snapshot.State = SessionLifecycleCompleted
		evidence.Reasons = []string{
			fmt.Sprintf("Last activity was %s, beyond the %s idle window.", FormatTimeAgo(lastActiveAt), FormatDuration(window.IdleWindow)),
			"No active marker remains for this session.",
			"This session is historical but still normal, so it is completed rather than stale.",
		}
	}

	evidence.State = snapshot.State
	snapshot.Evidence = evidence
	return snapshot
}

func activeCoherenceReason(hasActiveMarker bool) string {
	if hasActiveMarker {
		return "The active marker is present and still coherent with recent activity."
	}
	return "Recent activity is current and there is no conflicting marker evidence."
}

func healthProblemReason(problem string) string {
	if strings.TrimSpace(problem) == "" {
		return "Its session record is incomplete or unreliable."
	}
	return problem
}

// LifecycleStateMatchesQuery checks whether a query refers to a lifecycle state term.
func LifecycleStateMatchesQuery(state SessionLifecycleState, query string) bool {
	q := strings.TrimSpace(strings.ToLower(query))
	if q == "" {
		return false
	}

	stateTerm := strings.ToLower(string(state))
	return strings.Contains(stateTerm, q) || strings.Contains(q, stateTerm)
}

// SummarizeLifecycleSessions counts sessions by lifecycle state.
func SummarizeLifecycleSessions(sessions []Session) LifecycleSummary {
	var summary LifecycleSummary
	for _, session := range sessions {
		summary.Total++
		switch session.Lifecycle.State {
		case SessionLifecycleActive:
			summary.Active++
		case SessionLifecycleIdle:
			summary.Idle++
		case SessionLifecycleCompleted:
			summary.Completed++
		case SessionLifecycleStale:
			summary.Stale++
		}
	}
	return summary
}
