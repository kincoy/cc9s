package claudefs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanSkillRootsRecognizesDirectoryAndSingleFileSkills(t *testing.T) {
	root := t.TempDir()

	dirSkill := filepath.Join(root, "dir-skill")
	if err := os.MkdirAll(dirSkill, 0o755); err != nil {
		t.Fatalf("mkdir dir skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dirSkill, "SKILL.md"), []byte("# Dir Skill\n\nDirectory summary.\n"), 0o644); err != nil {
		t.Fatalf("write dir skill: %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, "single-skill.md"), []byte("# Single Skill\n\nSingle file summary.\n"), 0o644); err != nil {
		t.Fatalf("write single skill: %v", err)
	}

	skills, err := scanSkillRoots([]SkillDiscoveryRoot{{Path: root, Source: SkillSourceProject, Kind: SkillKindSkill}})
	if err != nil {
		t.Fatalf("scan skills: %v", err)
	}

	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}

	var foundDir, foundSingle bool
	for _, skill := range skills {
		switch skill.Name {
		case "dir-skill":
			foundDir = true
			if skill.Shape != SkillShapeDirectory {
				t.Fatalf("expected directory shape, got %s", skill.Shape)
			}
			if skill.Status != SkillStatusReady {
				t.Fatalf("expected ready status, got %s", skill.Status)
			}
		case "single-skill":
			foundSingle = true
			if skill.Shape != SkillShapeSingleFile {
				t.Fatalf("expected single-file shape, got %s", skill.Shape)
			}
			if skill.Status != SkillStatusReady {
				t.Fatalf("expected ready status, got %s", skill.Status)
			}
		}
	}

	if !foundDir || !foundSingle {
		t.Fatalf("expected both directory and single-file skills to be discovered")
	}
}

func TestScanSkillRootsKeepsInvalidDirectorySkill(t *testing.T) {
	root := t.TempDir()

	invalid := filepath.Join(root, "broken-skill")
	if err := os.MkdirAll(invalid, 0o755); err != nil {
		t.Fatalf("mkdir invalid skill: %v", err)
	}

	skills, err := scanSkillRoots([]SkillDiscoveryRoot{{Path: root, Source: SkillSourceUser, Kind: SkillKindSkill}})
	if err != nil {
		t.Fatalf("scan skills: %v", err)
	}

	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}

	if skills[0].Status != SkillStatusInvalid {
		t.Fatalf("expected invalid status, got %s", skills[0].Status)
	}
	if len(skills[0].ValidationReasons) == 0 {
		t.Fatalf("expected validation reasons for invalid skill")
	}
}

func TestScanSkillRootsAllowsDuplicateNamesAcrossSources(t *testing.T) {
	rootA := t.TempDir()
	rootB := t.TempDir()

	writeSingleSkill := func(root string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, "shared.md"), []byte("# Shared\n\nSummary.\n"), 0o644); err != nil {
			t.Fatalf("write skill: %v", err)
		}
	}

	writeSingleSkill(rootA)
	writeSingleSkill(rootB)

	skills, err := scanSkillRoots([]SkillDiscoveryRoot{
		{Path: rootA, Source: SkillSourceUser, Kind: SkillKindSkill},
		{Path: rootB, Source: SkillSourceProject, Kind: SkillKindSkill},
	})
	if err != nil {
		t.Fatalf("scan skills: %v", err)
	}

	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
	if skills[0].Source == skills[1].Source {
		t.Fatalf("expected duplicate names across different sources to remain distinct")
	}
}

func TestScanSkillRootsDeduplicatesIdenticalRootPaths(t *testing.T) {
	root := t.TempDir()
	roots := uniqueSkillDiscoveryRoots([]SkillDiscoveryRoot{
		{Path: root, Source: SkillSourceUser, Kind: SkillKindSkill},
		{Path: filepath.Join(root, "."), Source: SkillSourceProject, Kind: SkillKindSkill, ProjectName: "dup-project"},
	})

	if len(roots) != 1 {
		t.Fatalf("expected duplicate roots to collapse to 1, got %d", len(roots))
	}
	if roots[0].Source != SkillSourceUser {
		t.Fatalf("expected first root to win and preserve user source, got %s", roots[0].Source)
	}
}

func TestScanCommandRootIgnoresDirectories(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "plan.md"), []byte("# Plan\n\nCommand summary.\n"), 0o644); err != nil {
		t.Fatalf("write command: %v", err)
	}

	skills, err := scanSkillRoots([]SkillDiscoveryRoot{{Path: root, Source: SkillSourceProject, Kind: SkillKindCommand}})
	if err != nil {
		t.Fatalf("scan command roots: %v", err)
	}

	if len(skills) != 1 {
		t.Fatalf("expected only markdown commands to be scanned, got %d", len(skills))
	}
	if skills[0].Kind != SkillKindCommand {
		t.Fatalf("expected command kind, got %s", skills[0].Kind)
	}
}

func TestExtractSkillSummaryPrefersDescriptionMetadata(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "skill.md")
	content := `---
description: Metadata summary
---
# Title

Body summary.
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write skill file: %v", err)
	}

	summary, _ := extractSkillSummary(path)
	if summary != "Metadata summary" {
		t.Fatalf("expected metadata summary, got %q", summary)
	}
}

func TestExtractSkillSummaryFallsBackToReadableParagraph(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "skill.md")
	content := `# Title

This is the first readable paragraph.

- list item
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write skill file: %v", err)
	}

	summary, _ := extractSkillSummary(path)
	if summary != "This is the first readable paragraph." {
		t.Fatalf("unexpected summary: %q", summary)
	}
}
