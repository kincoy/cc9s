package claudefs

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

const (
	agentSectionCustom  = "custom"
	agentSectionPlugin  = "plugin"
	agentSectionBuiltIn = "builtin"
)

var runClaudeAgentsCommand = defaultRunClaudeAgentsCommand

var activeAgentsLinePattern = regexp.MustCompile(`^\d+\s+active agents$`)

func defaultRunClaudeAgentsCommand(workDir string, settingSources string) (string, error) {
	args := []string{"agents"}
	if strings.TrimSpace(settingSources) != "" {
		args = append(args, "--setting-sources", settingSources)
	}

	cmd := exec.Command("claude", args...)
	if strings.TrimSpace(workDir) != "" {
		cmd.Dir = workDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("claude agents failed: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	return string(output), nil
}

func lookupAgentRecognition(workDir string, settingSources string) (AgentRecognitionSnapshot, error) {
	output, err := runClaudeAgentsCommand(workDir, settingSources)
	if err != nil {
		return AgentRecognitionSnapshot{}, err
	}
	return parseAgentRecognitionOutput(output)
}

func parseAgentRecognitionOutput(output string) (AgentRecognitionSnapshot, error) {
	snapshot := AgentRecognitionSnapshot{
		CustomAgents: make(map[string]struct{}),
		PluginAgents: make(map[string]struct{}),
	}

	sawHeader := false
	section := ""
	for _, rawLine := range strings.Split(output, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		switch {
		case activeAgentsLinePattern.MatchString(line):
			sawHeader = true
			continue
		case strings.EqualFold(line, "Custom agents:"):
			if !sawHeader {
				return AgentRecognitionSnapshot{}, fmt.Errorf("invalid claude agents output: missing header")
			}
			section = agentSectionCustom
			continue
		case strings.EqualFold(line, "Plugin agents:"):
			if !sawHeader {
				return AgentRecognitionSnapshot{}, fmt.Errorf("invalid claude agents output: missing header")
			}
			section = agentSectionPlugin
			continue
		case strings.EqualFold(line, "Built-in agents:"):
			if !sawHeader {
				return AgentRecognitionSnapshot{}, fmt.Errorf("invalid claude agents output: missing header")
			}
			section = agentSectionBuiltIn
			continue
		}

		if normalizedSection, ok := normalizeAgentSectionHeading(line); ok {
			if !sawHeader {
				return AgentRecognitionSnapshot{}, fmt.Errorf("invalid claude agents output: missing header")
			}
			section = normalizedSection
			continue
		}

		if !sawHeader {
			return AgentRecognitionSnapshot{}, fmt.Errorf("invalid claude agents output: unexpected content %q", line)
		}

		if strings.HasSuffix(line, ":") {
			return AgentRecognitionSnapshot{}, fmt.Errorf("invalid claude agents output: unknown section %q", line)
		}

		if section == "" {
			return AgentRecognitionSnapshot{}, fmt.Errorf("invalid claude agents output: entry outside section %q", line)
		}

		name := parseRecognizedAgentName(line)
		if name == "" {
			return AgentRecognitionSnapshot{}, fmt.Errorf("invalid claude agents output: malformed agent entry %q", line)
		}

		switch section {
		case agentSectionCustom:
			snapshot.CustomAgents[name] = struct{}{}
		case agentSectionPlugin:
			snapshot.PluginAgents[name] = struct{}{}
		}
	}

	if !sawHeader {
		return AgentRecognitionSnapshot{}, fmt.Errorf("invalid claude agents output: missing active agents header")
	}

	return snapshot, nil
}

func parseRecognizedAgentName(line string) string {
	parts := strings.SplitN(strings.TrimSpace(line), " · ", 2)
	return strings.TrimSpace(parts[0])
}

func normalizeAgentSectionHeading(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasSuffix(trimmed, "agents:") {
		return "", false
	}

	lower := strings.ToLower(trimmed)
	switch lower {
	case "user agents:", "project agents:", "local agents:":
		return agentSectionCustom, true
	default:
		return "", false
	}
}

func agentRecognitionKey(agent AgentResource) string {
	if agent.Source == AgentSourcePlugin {
		if agent.PluginName == "" {
			return ""
		}
		return agent.PluginName + ":" + agent.Name
	}
	return agent.Name
}

func agentRecognizedBySnapshot(agent AgentResource, snapshot AgentRecognitionSnapshot) bool {
	key := agentRecognitionKey(agent)
	if key == "" {
		return false
	}

	if agent.Source == AgentSourcePlugin {
		_, ok := snapshot.PluginAgents[key]
		return ok
	}

	_, ok := snapshot.CustomAgents[key]
	return ok
}
