package ui

import (
	"testing"

	"github.com/kincoy/cc9s/internal/claudefs"
)

func TestPreferredEditorPriority(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")
	if got := preferredEditor(); got != "vim" {
		t.Fatalf("preferredEditor() = %q, want vim", got)
	}

	t.Setenv("EDITOR", "nvim")
	if got := preferredEditor(); got != "nvim" {
		t.Fatalf("preferredEditor() with EDITOR = %q, want nvim", got)
	}

	t.Setenv("VISUAL", "hx")
	if got := preferredEditor(); got != "hx" {
		t.Fatalf("preferredEditor() with VISUAL = %q, want hx", got)
	}
}

func TestSkillEditTargetPath(t *testing.T) {
	dirSkill := claudefs.SkillResource{
		Name:      "dir-skill",
		Path:      "/tmp/dir-skill",
		EntryFile: "/tmp/dir-skill/SKILL.md",
		Shape:     claudefs.SkillShapeDirectory,
	}
	target, err := skillEditTargetPath(dirSkill)
	if err != nil {
		t.Fatalf("skillEditTargetPath(dir) error: %v", err)
	}
	if target != "/tmp/dir-skill/SKILL.md" {
		t.Fatalf("directory target = %q, want entry file", target)
	}

	fileSkill := claudefs.SkillResource{
		Name:      "file-skill",
		Path:      "/tmp/file-skill.md",
		EntryFile: "/tmp/file-skill.md",
		Shape:     claudefs.SkillShapeSingleFile,
	}
	target, err = skillEditTargetPath(fileSkill)
	if err != nil {
		t.Fatalf("skillEditTargetPath(file) error: %v", err)
	}
	if target != "/tmp/file-skill.md" {
		t.Fatalf("single-file target = %q, want original file", target)
	}
}

func TestBuildSkillEditorCommandUsesConfiguredEditor(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "nvim -u NONE")

	skill := claudefs.SkillResource{
		Name:      "skill",
		Path:      "/tmp/skill",
		EntryFile: "/tmp/skill/SKILL.md",
		Shape:     claudefs.SkillShapeDirectory,
	}

	cmd, err := buildSkillEditorCommand(skill, skill.EntryFile)
	if err != nil {
		t.Fatalf("buildSkillEditorCommand error: %v", err)
	}

	if got := cmd.Args[0]; got != "nvim" {
		t.Fatalf("editor command = %q, want nvim", got)
	}
	if got := cmd.Args[len(cmd.Args)-1]; got != skill.EntryFile {
		t.Fatalf("editor target = %q, want %q", got, skill.EntryFile)
	}
	if got := cmd.Dir; got != skill.Path {
		t.Fatalf("working dir = %q, want %q", got, skill.Path)
	}
}

func TestBuildSkillEditorCommandPreservesQuotedExecutable(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "\"/Applications/Vim With Space/bin/vim\" -u NONE")

	skill := claudefs.SkillResource{
		Name:      "skill",
		Path:      "/tmp/skill",
		EntryFile: "/tmp/skill/SKILL.md",
		Shape:     claudefs.SkillShapeDirectory,
	}

	cmd, err := buildSkillEditorCommand(skill, skill.EntryFile)
	if err != nil {
		t.Fatalf("buildSkillEditorCommand error: %v", err)
	}

	if got := cmd.Args[0]; got != "/Applications/Vim With Space/bin/vim" {
		t.Fatalf("editor command = %q, want quoted executable preserved", got)
	}
	if got := cmd.Args[1]; got != "-u" {
		t.Fatalf("expected first argument after executable to be -u, got %q", got)
	}
}

func TestSplitCommandLineRejectsUnterminatedQuotes(t *testing.T) {
	_, err := splitCommandLine(`"nvim -u NONE`)
	if err == nil {
		t.Fatal("expected unterminated quote error")
	}
}

func TestSkillEditTargetPathRejectsEmptySkill(t *testing.T) {
	_, err := skillEditTargetPath(claudefs.SkillResource{})
	if err == nil {
		t.Fatal("expected error for empty skill target")
	}
}
