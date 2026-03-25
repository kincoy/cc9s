package cli

import (
	"fmt"
	"os"
)

// renderJSONMode dispatches JSON rendering for each result type.
func renderJSONMode(w *os.File, result CommandResult) {
	switch r := result.(type) {
	case StatusResult:
		mustRenderJSON(w, r)
	case ProjectListResult:
		mustRenderJSON(w, r.Projects)
	case ProjectDetailResult:
		mustRenderJSON(w, r.ProjectDetail)
	case SessionListResult:
		mustRenderJSON(w, r.Sessions)
	case SessionDetailResult:
		mustRenderJSON(w, r.SessionDetail)
	case SkillListResult:
		renderSkillListJSON(w, r)
	case AgentListResult:
		renderAgentListJSON(w, r)
	case AgentDetailResult:
		mustRenderJSON(w, r.AgentDetail)
	case CleanupResult:
		mustRenderJSON(w, r)
	case HelpResult:
		_, _ = w.WriteString(r.Text)
	case VersionResult:
		mustRenderJSON(w, map[string]string{"version": r.Version})
	default:
		mustRenderJSON(w, map[string]string{"error": "unknown result type"})
	}
}

func mustRenderJSON(w *os.File, v interface{}) {
	if err := renderJSON(w, v); err != nil {
		fmt.Fprintf(os.Stderr, "error: json output: %v\n", err)
	}
}

// renderSkillListJSON renders skill list as a JSON array.
// Valid skills omit validation_reasons per contract.
func renderSkillListJSON(w *os.File, r SkillListResult) {
	out := make([]interface{}, 0, len(r.Skills))
	for _, s := range r.Skills {
		if s.ValidationReasons == nil {
			out = append(out, map[string]interface{}{
				"name":    s.Name,
				"type":    s.Type,
				"scope":   s.Scope,
				"status":  s.Status,
				"project": s.Project,
				"path":    s.Path,
			})
		} else {
			out = append(out, s)
		}
	}
	mustRenderJSON(w, out)
}

// renderAgentListJSON renders agent list as a JSON array.
// Valid agents omit validation_reasons per contract.
func renderAgentListJSON(w *os.File, r AgentListResult) {
	out := make([]interface{}, 0, len(r.Agents))
	for _, a := range r.Agents {
		if a.ValidationReasons == nil {
			out = append(out, map[string]interface{}{
				"name":    a.Name,
				"scope":   a.Scope,
				"status":  a.Status,
				"project": a.Project,
				"path":    a.Path,
			})
		} else {
			out = append(out, a)
		}
	}
	mustRenderJSON(w, out)
}
