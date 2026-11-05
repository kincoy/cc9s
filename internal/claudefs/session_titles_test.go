package claudefs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSessionJSONL(t *testing.T, lines ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "session.jsonl")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatalf("write session file: %v", err)
	}
	return path
}

func TestSessionTitlesDisplayNamePrecedence(t *testing.T) {
	tests := []struct {
		name   string
		titles SessionTitles
		want   string
	}{
		{
			name:   "custom title wins",
			titles: SessionTitles{CustomTitle: "邮件二期", AgentName: "agent", AITitle: "ai"},
			want:   "邮件二期",
		},
		{
			name:   "agent name beats ai title",
			titles: SessionTitles{AgentName: "agent", AITitle: "ai"},
			want:   "agent",
		},
		{
			name:   "ai title fallback",
			titles: SessionTitles{AITitle: "ai"},
			want:   "ai",
		},
		{
			name:   "no titles",
			titles: SessionTitles{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.titles.DisplayName(); got != tt.want {
				t.Fatalf("DisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractSessionTitles(t *testing.T) {
	path := writeSessionJSONL(t,
		`{"type":"ai-title","aiTitle":"MRD 需求实现方案"}`,
		`{"type":"agent-name","agentName":"邮件二期"}`,
		`{"type":"custom-title","customTitle":"邮件二期"}`,
		`{"type":"custom-title","customTitle":"最终名称"}`,
	)

	got := ExtractSessionTitles(path)
	if got.CustomTitle != "最终名称" {
		t.Fatalf("CustomTitle = %q, want latest value", got.CustomTitle)
	}
	if got.AgentName != "邮件二期" {
		t.Fatalf("AgentName = %q", got.AgentName)
	}
	if got.AITitle != "MRD 需求实现方案" {
		t.Fatalf("AITitle = %q", got.AITitle)
	}
	if got.DisplayName() != "最终名称" {
		t.Fatalf("DisplayName() = %q", got.DisplayName())
	}
}

func TestExtractSessionTitlesReadsFileTail(t *testing.T) {
	padding := fmt.Sprintf(`{"type":"user","message":{"content":"%s"}}`, strings.Repeat("x", 70*1024))
	path := writeSessionJSONL(t,
		padding,
		`{"type":"custom-title","customTitle":"尾部标题"}`,
	)

	got := ExtractSessionTitles(path)
	if got.CustomTitle != "尾部标题" {
		t.Fatalf("CustomTitle = %q, want title from file tail", got.CustomTitle)
	}
}

func TestExtractSessionTitlesMissingFile(t *testing.T) {
	got := ExtractSessionTitles(filepath.Join(t.TempDir(), "missing.jsonl"))
	if got.DisplayName() != "" {
		t.Fatalf("DisplayName() = %q, want empty for missing file", got.DisplayName())
	}
}

func TestCleanTitle(t *testing.T) {
	if got := cleanTitle("  邮件\t二期\n "); got != "邮件 二期" {
		t.Fatalf("cleanTitle() = %q, want %q", got, "邮件 二期")
	}
	if got := cleanTitle(nil); got != "" {
		t.Fatalf("cleanTitle(nil) = %q, want empty", got)
	}
	if got := cleanTitle(42); got != "" {
		t.Fatalf("cleanTitle(42) = %q, want empty", got)
	}
}

func TestLoadProjectSessionsIncludesSessionName(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	encodedPath := "-tmp-cc9s-demo"
	projectDir := filepath.Join(home, ".claude", "projects", encodedPath)
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}

	content := strings.Join([]string{
		`{"type":"user","sessionId":"sess-1","cwd":"/tmp/cc9s/demo","version":"2.0.0","message":{"content":"hello"}}`,
		`{"type":"ai-title","aiTitle":"AI 标题"}`,
		`{"type":"custom-title","customTitle":"演示会话"}`,
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(projectDir, "sess-1.jsonl"), []byte(content), 0o644); err != nil {
		t.Fatalf("write session: %v", err)
	}

	sessions, err := LoadProjectSessions(encodedPath)
	if err != nil {
		t.Fatalf("LoadProjectSessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("got %d sessions, want 1", len(sessions))
	}
	if sessions[0].Name != "演示会话" {
		t.Fatalf("session name = %q, want 演示会话", sessions[0].Name)
	}
}
