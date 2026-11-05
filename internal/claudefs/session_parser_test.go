package claudefs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tool results in real transcripts routinely exceed bufio.Scanner's 64 KiB
// default; parsing must not fail on them.
func TestParseSessionStatsHandlesLongLines(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	projectDir := filepath.Join(home, ".claude", "projects", "-tmp-demo")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}

	longText := strings.Repeat("x", 256*1024)
	content := strings.Join([]string{
		`{"type":"user","timestamp":"2026-01-01T00:00:00Z","sessionId":"sess-long","message":{"role":"user","content":"` + longText + `"}}`,
		`{"type":"assistant","timestamp":"2026-01-01T00:00:10Z","message":{"model":"claude-test","usage":{"input_tokens":10,"output_tokens":5,"cache_creation_input_tokens":2,"cache_read_input_tokens":3},"content":[{"type":"tool_use","name":"Read"}]}}`,
	}, "\n") + "\n"

	if err := os.WriteFile(filepath.Join(projectDir, "sess-long.jsonl"), []byte(content), 0o644); err != nil {
		t.Fatalf("write session: %v", err)
	}

	stats, err := ParseSessionStats(Session{ID: "sess-long"})
	if err != nil {
		t.Fatalf("ParseSessionStats: %v", err)
	}
	if stats.Model != "claude-test" {
		t.Fatalf("model = %q, want claude-test", stats.Model)
	}
	if stats.UserMsgCount != 1 || stats.ToolCallCount != 1 {
		t.Fatalf("unexpected stats: user=%d tools=%d", stats.UserMsgCount, stats.ToolCallCount)
	}

	entries, total, err := ParseSessionLog("sess-long", 0, 10)
	if err != nil {
		t.Fatalf("ParseSessionLog: %v", err)
	}
	if total != 1 || len(entries) != 1 {
		t.Fatalf("log total=%d entries=%d, want 1/1", total, len(entries))
	}
	if len(entries[0].ToolCalls) != 1 {
		t.Fatalf("tool calls = %d, want 1", len(entries[0].ToolCalls))
	}
}
