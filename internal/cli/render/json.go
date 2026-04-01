package render

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kincoy/cc9s/internal/cli/contract"
)

// WriteResult renders a command result as either text or JSON.
func WriteResult(stdout, stderr io.Writer, mode contract.OutputMode, result contract.Result) error {
	if mode == contract.OutputJSON {
		return writeJSONResult(stdout, result)
	}
	return writeTextResult(stdout, stderr, result)
}

func writeJSONResult(stdout io.Writer, result contract.Result) error {
	switch r := result.(type) {
	case contract.StatusResult:
		return writeJSON(stdout, r)
	case contract.ProjectListResult:
		return writeJSON(stdout, r.Projects)
	case contract.ProjectDetailResult:
		return writeJSON(stdout, r.ProjectDetail)
	case contract.SessionListResult:
		return writeJSON(stdout, r.Sessions)
	case contract.SessionDetailResult:
		return writeJSON(stdout, r.SessionDetail)
	case contract.SkillListResult:
		return writeSkillListJSON(stdout, r)
	case contract.AgentListResult:
		return writeAgentListJSON(stdout, r)
	case contract.AgentDetailResult:
		return writeJSON(stdout, r.AgentDetail)
	case contract.ThemesResult:
		return writeJSON(stdout, r.Themes)
	case contract.CleanupResult:
		return writeJSON(stdout, r)
	case contract.VersionResult:
		return writeJSON(stdout, map[string]string{"version": r.Version})
	default:
		return writeJSON(stdout, contract.ErrorPayload{Error: "unknown result type"})
	}
}

func writeSkillListJSON(stdout io.Writer, r contract.SkillListResult) error {
	out := make([]any, 0, len(r.Skills))
	for _, s := range r.Skills {
		if s.ValidationReasons == nil {
			out = append(out, map[string]any{
				"name":    s.Name,
				"type":    s.Type,
				"scope":   s.Scope,
				"status":  s.Status,
				"project": s.Project,
				"path":    s.Path,
			})
			continue
		}
		out = append(out, s)
	}
	return writeJSON(stdout, out)
}

func writeAgentListJSON(stdout io.Writer, r contract.AgentListResult) error {
	out := make([]any, 0, len(r.Agents))
	for _, a := range r.Agents {
		if a.ValidationReasons == nil {
			out = append(out, map[string]any{
				"name":    a.Name,
				"scope":   a.Scope,
				"status":  a.Status,
				"project": a.Project,
				"path":    a.Path,
			})
			continue
		}
		out = append(out, a)
	}
	return writeJSON(stdout, out)
}

func writeJSON(stdout io.Writer, v any) error {
	if stdout == nil {
		return fmt.Errorf("nil stdout writer")
	}

	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	if _, err := stdout.Write(data); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	if _, err := io.WriteString(stdout, "\n"); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}
