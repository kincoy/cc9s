package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
)

func editSkillCmd(skill claudefs.SkillResource) tea.Cmd {
	return func() tea.Msg {
		targetPath, err := skillEditTargetPath(skill)
		if err != nil {
			return SkillEditorFinishedMsg{Skill: skill, Err: err}
		}

		cmd, err := buildSkillEditorCommand(skill, targetPath)
		if err != nil {
			return SkillEditorFinishedMsg{Skill: skill, Err: err}
		}

		clearScreen()
		return tea.ExecProcess(cmd, func(err error) tea.Msg {
			clearScreen()
			return SkillEditorFinishedMsg{Skill: skill, Err: err}
		})()
	}
}

func buildSkillEditorCommand(skill claudefs.SkillResource, targetPath string) (*exec.Cmd, error) {
	editor := preferredEditor()
	parts, err := splitCommandLine(editor)
	if err != nil {
		return nil, err
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("no editor configured")
	}

	cmd := exec.Command(parts[0], append(parts[1:], targetPath)...)
	cmd.Dir = skillWorkingDir(skill, targetPath)
	return cmd, nil
}

func preferredEditor() string {
	if visual := strings.TrimSpace(os.Getenv("VISUAL")); visual != "" {
		return visual
	}
	if editor := strings.TrimSpace(os.Getenv("EDITOR")); editor != "" {
		return editor
	}
	return "vim"
}

func skillEditTargetPath(skill claudefs.SkillResource) (string, error) {
	target := strings.TrimSpace(skill.EntryFile)
	if target == "" {
		target = strings.TrimSpace(skill.Path)
	}
	if target == "" {
		return "", fmt.Errorf("skill %q has no editable file", skill.Name)
	}
	return target, nil
}

func skillWorkingDir(skill claudefs.SkillResource, targetPath string) string {
	if skill.Shape == claudefs.SkillShapeDirectory && skill.Path != "" {
		return skill.Path
	}
	return filepath.Dir(targetPath)
}

func splitCommandLine(input string) ([]string, error) {
	var (
		args         []string
		current      strings.Builder
		inSingle     bool
		inDouble     bool
		escapeNext   bool
		sawCharacter bool
	)

	flush := func() {
		if !sawCharacter {
			return
		}
		args = append(args, current.String())
		current.Reset()
		sawCharacter = false
	}

	for _, r := range input {
		switch {
		case escapeNext:
			current.WriteRune(r)
			sawCharacter = true
			escapeNext = false
		case r == '\\' && !inSingle:
			escapeNext = true
		case r == '\'' && !inDouble:
			inSingle = !inSingle
			sawCharacter = true
		case r == '"' && !inSingle:
			inDouble = !inDouble
			sawCharacter = true
		case unicode.IsSpace(r) && !inSingle && !inDouble:
			flush()
		default:
			current.WriteRune(r)
			sawCharacter = true
		}
	}

	if escapeNext {
		current.WriteRune('\\')
	}
	if inSingle || inDouble {
		return nil, fmt.Errorf("unterminated quote in editor command")
	}

	flush()
	return args, nil
}
