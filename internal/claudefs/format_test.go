package claudefs

import (
	"os"
	"testing"
)

func TestDecodePathFS(t *testing.T) {
	tests := []struct {
		name     string
		encoded  string
		expected string
	}{
		// Simple path (no ambiguity)
		{"simple path", "-Users-kinco", "/Users/kinco"},

		// Hidden directories (empty segment = ".")
		{"hidden .claude", "-Users-kinco--claude", "/Users/kinco/.claude"},
		{"hidden .config", "-Users-kinco--config-sketchybar", "/Users/kinco/.config/sketchybar"},
		{"hidden .local", "-Users-kinco--local-share-chezmoi", "/Users/kinco/.local/share/chezmoi"},

		// github.com paths ("." encoding ambiguity)
		{"github.com repo", "-Users-kinco-go-src-github-com-kincoy-cc9s", "/Users/kinco/go/src/github.com/kincoy/cc9s"},
		{"github.com todolist", "-Users-kinco-go-src-github-com-kincoy-todolist", "/Users/kinco/go/src/github.com/kincoy/todolist"},
		{"github.com kite-org", "-Users-kinco-go-src-github-com-kite-org-kite", "/Users/kinco/go/src/github.com/kite-org/kite"},

		// Directory names with dashes ("-" encoding ambiguity)
		{"affaan-m", "-Users-kinco-go-src-github-com-affaan-m-everything-claude-code", "/Users/kinco/go/src/github.com/affaan-m/everything-claude-code"},
		{"kagent-dev", "-Users-kinco-go-src-github-com-kagent-dev-kagent", "/Users/kinco/go/src/github.com/kagent-dev/kagent"},
		{"kubernetes-sigs", "-Users-kinco-go-src-github-com-kubernetes-sigs-mcp-lifecycle-operator", "/Users/kinco/go/src/github.com/kubernetes-sigs/mcp-lifecycle-operator"},

		// k8s.io paths
		{"k8s.io", "-Users-kinco-go-src-k8s-io-autoscaler", "/Users/kinco/go/src/k8s.io/autoscaler"},
		{"k8s.io cluster-autoscaler", "-Users-kinco-go-src-k8s-io-autoscaler-cluster-autoscaler", "/Users/kinco/go/src/k8s.io/autoscaler/cluster-autoscaler"},
		{"k8s.io kubernetes", "-Users-kinco-go-src-k8s-io-kubernetes", "/Users/kinco/go/src/k8s.io/kubernetes"},
		{"sigs.k8s.io", "-Users-kinco-go-src-sigs-k8s-io-kwok", "/Users/kinco/go/src/sigs.k8s.io/kwok"},

		// Mixed separator domains (domain contains both "-" and ".", e.g. example-go.com)
		{"mixed-sep domain (dash+dot)", "-Users-kinco-go-src-thinkingdata-go-com-cloud-native-helm-charts", "/Users/kinco/go/src/thinkingdata-go.com/cloud-native/helm-charts"},

		// Custom .cn domain path
		{".cn domain path", "-Users-kinco-workspace-repos-thinkingdata-cn-refine-etchosts", "/Users/kinco/workspace/repos/thinkingdata.cn/refine-etchosts"},

		// github.com paths under workspace
		{"workspace github.com cli", "-Users-kinco-workspace-repos-github-com-cli", "/Users/kinco/workspace/repos/github.com/cli"},
		{"workspace github.com nvimdots", "-Users-kinco-workspace-repos-github-com-nvimdots", "/Users/kinco/workspace/repos/github.com/nvimdots"},

		// Complex path: hidden directory + dashes + multi-level domain (hardest case)
		{"complex: hidden+dash+domain+worktree", "-Users-kinco-workspace-repos-thinkingdata-cn-ta-admin-service--claude-worktrees-support-sd", "/Users/kinco/workspace/repos/thinkingdata.cn/ta-admin-service/.claude/worktrees/support-sd"},

		// /private/tmp path
		{"private tmp", "-private-tmp", "/private/tmp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if the target directory does not exist (avoid errors on other machines)
			if _, err := os.Stat(tt.expected); err != nil {
				t.Skipf("target directory does not exist: %s", tt.expected)
			}

			got := DecodePathFS(tt.encoded)
			if got != tt.expected {
				t.Errorf("DecodePathFS(%q) =\n  %q,\nwant %q", tt.encoded, got, tt.expected)
			}
		})
	}
}

func TestBuildCandidates(t *testing.T) {
	tests := []struct {
		segments  []string
		expectLen int
		// Check whether specific candidates are present
		contains []string
	}{
		{
			segments:  []string{"github", "com"},
			expectLen: 2,
			contains:  []string{"github.com", "github-com"},
		},
		{
			segments:  []string{"ta", "admin", "service"},
			expectLen: 4,
			contains:  []string{"ta-admin-service", "ta.admin.service", "ta.admin-service", "ta-admin.service"},
		},
		{
			segments:  []string{"", "claude"},
			expectLen: 2,
			contains:  []string{".claude", "-claude"},
		},
		{
			segments:  []string{"example", "go", "com"},
			expectLen: 4,
			contains:  []string{"example-go.com", "example.go.com", "example-go-com", "example.go-com"},
		},
		{
			segments:  []string{"support", "sd"},
			expectLen: 2,
			contains:  []string{"support-sd", "support.sd"},
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := buildCandidates(tt.segments)
			if len(got) != tt.expectLen {
				t.Errorf("buildCandidates(%v) got %d candidates, want %d: %v", tt.segments, len(got), tt.expectLen, got)
			}
			for _, want := range tt.contains {
				found := false
				for _, c := range got {
					if c == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("buildCandidates(%v) missing candidate %q, got %v", tt.segments, want, got)
				}
			}
		})
	}
}

func TestDecodePathFSEmpty(t *testing.T) {
	got := DecodePathFS("")
	if got != "" {
		t.Errorf("DecodePathFS(\"\") = %q, want \"\"", got)
	}
}

func TestDecodePathFSRoot(t *testing.T) {
	// "-" should decode to "/"
	got := DecodePathFS("-")
	if got != "/" {
		t.Errorf("DecodePathFS(\"-\") = %q, want \"/\"", got)
	}
}
