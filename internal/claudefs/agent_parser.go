package claudefs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type agentFrontmatter struct {
	Name           string      `yaml:"name"`
	Description    string      `yaml:"description"`
	Tools          interface{} `yaml:"tools"`
	Model          string      `yaml:"model"`
	PermissionMode string      `yaml:"permissionMode"`
	Memory         string      `yaml:"memory"`
}

func parseAgentFile(root AgentDiscoveryRoot, filePath string) AgentResource {
	nameFromPath := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	resource := AgentResource{
		Name:              nameFromPath,
		Path:              filePath,
		Source:            root.Source,
		Scope:             agentSourceScope(root.Source),
		ProjectName:       root.ProjectName,
		ProjectPath:       root.ProjectPath,
		PluginName:        root.PluginName,
		PluginInstallMode: root.PluginInstallMode,
		Status:            AgentStatusInvalid,
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		resource.ValidationReasons = normalizedAgentReasons(resource.Status, []string{"Agent definition file is not readable"})
		resource.Availability = AgentAvailabilityResult{Status: resource.Status, Reasons: resource.ValidationReasons}
		return resource
	}

	content := string(data)
	var reasons []string
	body := content
	frontmatter, body, frontmatterFound, parseErr := parseAgentFrontmatter(content)
	if parseErr != nil {
		reasons = append(reasons, "Agent frontmatter is unreadable or malformed")
	} else if !frontmatterFound {
		reasons = append(reasons, "Missing required identifying metadata")
	} else {
		if trimmedName := strings.TrimSpace(frontmatter.Name); trimmedName != "" {
			resource.Name = trimmedName
		} else {
			reasons = append(reasons, "Missing required identifying metadata")
		}

		resource.Description = strings.TrimSpace(frontmatter.Description)
		resource.Model = strings.TrimSpace(frontmatter.Model)
		resource.PermissionMode = strings.TrimSpace(frontmatter.PermissionMode)
		resource.Memory = strings.TrimSpace(frontmatter.Memory)
		resource.Tools = parseAgentTools(frontmatter.Tools)
	}

	resource.Summary, reasons = extractAgentSummary(resource.Description, body, reasons)
	resource.ValidationReasons = normalizedAgentReasons(resource.Status, reasons)
	resource.Availability = AgentAvailabilityResult{Status: resource.Status, Reasons: resource.ValidationReasons}
	return resource
}

func parseAgentFrontmatter(content string) (agentFrontmatter, string, bool, error) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return agentFrontmatter{}, content, false, nil
	}

	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return agentFrontmatter{}, content, true, fmt.Errorf("missing closing frontmatter delimiter")
	}

	var meta agentFrontmatter
	frontmatterContent := strings.Join(lines[1:end], "\n")
	if err := yaml.Unmarshal([]byte(frontmatterContent), &meta); err != nil {
		return agentFrontmatter{}, content, true, err
	}

	body := strings.Join(lines[end+1:], "\n")
	return meta, body, true, nil
}

func parseAgentTools(raw interface{}) []string {
	switch v := raw.(type) {
	case nil:
		return nil
	case []interface{}:
		tools := make([]string, 0, len(v))
		for _, item := range v {
			if text := strings.TrimSpace(fmt.Sprint(item)); text != "" && text != "<nil>" {
				tools = append(tools, text)
			}
		}
		return tools
	case []string:
		tools := make([]string, 0, len(v))
		for _, item := range v {
			if text := strings.TrimSpace(item); text != "" {
				tools = append(tools, text)
			}
		}
		return tools
	default:
		if text := strings.TrimSpace(fmt.Sprint(v)); text != "" && text != "<nil>" {
			return []string{text}
		}
		return nil
	}
}

func extractAgentSummary(description, body string, reasons []string) (string, []string) {
	if description = strings.TrimSpace(description); description != "" {
		reasons = append(reasons, "Summary derived from description metadata")
		return description, reasons
	}

	if paragraph := extractFirstReadableParagraph(body); paragraph != "" {
		reasons = append(reasons, "Summary derived from markdown description")
		return paragraph, reasons
	}

	return "", reasons
}

func agentSourceScope(source AgentSource) AgentScope {
	switch source {
	case AgentSourceProject:
		return AgentScopeProject
	case AgentSourcePlugin:
		return AgentScopePlugin
	default:
		return AgentScopeUser
	}
}

func normalizedAgentReasons(status AgentStatus, reasons []string) []string {
	clean := make([]string, 0, len(reasons))
	seen := make(map[string]struct{}, len(reasons))
	for _, reason := range reasons {
		reason = strings.TrimSpace(reason)
		if reason == "" {
			continue
		}
		if _, ok := seen[reason]; ok {
			continue
		}
		seen[reason] = struct{}{}
		clean = append(clean, reason)
	}

	if len(clean) == 0 {
		if status == AgentStatusReady {
			return []string{"Agent is recognized by Claude Code"}
		}
		return []string{"Agent file is not recognized by Claude Code"}
	}

	return clean
}
