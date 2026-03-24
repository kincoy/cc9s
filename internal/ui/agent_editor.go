package ui

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/kincoy/cc9s/internal/claudefs"
)

func editAgentCmd(agent claudefs.AgentResource) tea.Cmd {
	return func() tea.Msg {
		targetPath, err := agentEditTargetPath(agent)
		if err != nil {
			return AgentEditorFinishedMsg{Agent: agent, Err: err}
		}

		cmd, err := buildAgentEditorCommand(agent, targetPath)
		if err != nil {
			return AgentEditorFinishedMsg{Agent: agent, Err: err}
		}

		clearScreen()
		return tea.ExecProcess(cmd, func(err error) tea.Msg {
			clearScreen()
			return AgentEditorFinishedMsg{Agent: agent, Err: err}
		})()
	}
}

func buildAgentEditorCommand(agent claudefs.AgentResource, targetPath string) (*exec.Cmd, error) {
	return buildFileEditorCommand(targetPath, agentWorkingDir(agent, targetPath))
}

func agentEditTargetPath(agent claudefs.AgentResource) (string, error) {
	target := strings.TrimSpace(agent.Path)
	if target == "" {
		return "", fmt.Errorf("agent %q has no editable file", agent.Name)
	}
	return target, nil
}

func agentWorkingDir(agent claudefs.AgentResource, targetPath string) string {
	return filepath.Dir(targetPath)
}
