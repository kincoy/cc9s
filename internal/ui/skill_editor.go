package ui

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

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
	return buildFileEditorCommand(targetPath, skillWorkingDir(skill, targetPath))
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
