package root

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestExecuteWritesTextErrors(t *testing.T) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	defer stdoutR.Close()
	defer stderrR.Close()
	defer stdoutW.Close()
	defer stderrW.Close()
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	code := Execute([]string{"sessions", "cleanup"})

	_ = stdoutW.Close()
	_ = stderrW.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if _, err := stdout.ReadFrom(stdoutR); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if _, err := stderr.ReadFrom(stderrR); err != nil {
		t.Fatalf("read stderr: %v", err)
	}

	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout should be empty, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "--dry-run is required") {
		t.Fatalf("stderr = %q, want cleanup guard error", stderr.String())
	}
}

func TestExecuteWritesJSONErrors(t *testing.T) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	defer stdoutR.Close()
	defer stderrR.Close()
	defer stdoutW.Close()
	defer stderrW.Close()
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	code := Execute([]string{"sessions", "cleanup", "--json"})

	_ = stdoutW.Close()
	_ = stderrW.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if _, err := stdout.ReadFrom(stdoutR); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if _, err := stderr.ReadFrom(stderrR); err != nil {
		t.Fatalf("read stderr: %v", err)
	}

	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if strings.TrimSpace(stdout.String()) != `{"error":"--dry-run is required for cleanup (preview-only in v1)"}` {
		t.Fatalf("stdout = %q, want JSON error payload", strings.TrimSpace(stdout.String()))
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty, got %q", stderr.String())
	}
}

func TestExecuteVersionFlagMatchesVersionCommand(t *testing.T) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	defer stdoutR.Close()
	defer stderrR.Close()
	defer stdoutW.Close()
	defer stderrW.Close()
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	code := Execute([]string{"--version", "--json"})

	_ = stdoutW.Close()
	_ = stderrW.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if _, err := stdout.ReadFrom(stdoutR); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if _, err := stderr.ReadFrom(stderrR); err != nil {
		t.Fatalf("read stderr: %v", err)
	}

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if strings.TrimSpace(stdout.String()) == "" {
		t.Fatalf("stdout should not be empty")
	}
	if !strings.Contains(strings.TrimSpace(stdout.String()), `"version":`) {
		t.Fatalf("stdout = %q, want JSON version payload", strings.TrimSpace(stdout.String()))
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty, got %q", stderr.String())
	}
}

func TestExecuteBareJSONReturnsJSONError(t *testing.T) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	defer stdoutR.Close()
	defer stderrR.Close()
	defer stdoutW.Close()
	defer stderrW.Close()
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	code := Execute([]string{"--json"})

	_ = stdoutW.Close()
	_ = stderrW.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if _, err := stdout.ReadFrom(stdoutR); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if _, err := stderr.ReadFrom(stderrR); err != nil {
		t.Fatalf("read stderr: %v", err)
	}

	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if strings.TrimSpace(stdout.String()) != `{"error":"expected a command"}` {
		t.Fatalf("stdout = %q, want JSON error payload", strings.TrimSpace(stdout.String()))
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty, got %q", stderr.String())
	}
}

func TestExecuteRootHelpIncludesAutomationGuidance(t *testing.T) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	oldWidth := os.Getenv("__FANG_TEST_WIDTH")
	_ = os.Setenv("__FANG_TEST_WIDTH", "120")
	defer func() {
		_ = os.Setenv("__FANG_TEST_WIDTH", oldWidth)
	}()
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	defer stdoutR.Close()
	defer stderrR.Close()
	defer stdoutW.Close()
	defer stderrW.Close()
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	code := Execute([]string{"--help"})

	_ = stdoutW.Close()
	_ = stderrW.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if _, err := stdout.ReadFrom(stdoutR); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if _, err := stderr.ReadFrom(stderrR); err != nil {
		t.Fatalf("read stderr: %v", err)
	}

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty, got %q", stderr.String())
	}
	for _, want := range []string{
		"Launch the interactive TUI when no command is provided.",
		"Add --json for machine-readable output.",
		"Use `cc9s <resource> --help` for resource-specific flags and enums.",
		"cc9s status --json",
		"cc9s sessions inspect <id> --json",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout.String())
		}
	}
}

func TestExecuteSessionsHelpIncludesShortcutAndCleanupGuidance(t *testing.T) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	oldWidth := os.Getenv("__FANG_TEST_WIDTH")
	_ = os.Setenv("__FANG_TEST_WIDTH", "120")
	defer func() {
		_ = os.Setenv("__FANG_TEST_WIDTH", oldWidth)
	}()
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	defer stdoutR.Close()
	defer stderrR.Close()
	defer stdoutW.Close()
	defer stderrW.Close()
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	code := Execute([]string{"sessions", "--help"})

	_ = stdoutW.Close()
	_ = stderrW.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if _, err := stdout.ReadFrom(stdoutR); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if _, err := stderr.ReadFrom(stderrR); err != nil {
		t.Fatalf("read stderr: %v", err)
	}

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty, got %q", stderr.String())
	}
	for _, want := range []string{
		"`cc9s sessions <id>` is an inspect shortcut.",
		"`cleanup` is preview-only and requires `--dry-run`.",
		"--sort: updated | state | project",
		"cc9s sessions cleanup --dry-run --older-than 7d",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout.String())
		}
	}
}
