package action

import (
	"strings"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/cli/contract"
)

// SkillsList returns the skills list projection.
func SkillsList(opts contract.SkillListOptions) (contract.SkillListResult, error) {
	skillResult := scanSkills()
	if skillResult.Err != nil {
		return contract.SkillListResult{}, statusError("scan skills: %v", skillResult.Err)
	}

	entries := make([]contract.SkillListEntry, 0, len(skillResult.Skills))
	for _, s := range skillResult.Skills {
		if opts.Project != "" && !strings.EqualFold(s.ProjectName, opts.Project) {
			continue
		}
		if opts.Scope != "" && !scopeMatches(string(s.Scope), opts.Scope) {
			continue
		}
		if opts.Type != "" && !scopeMatches(string(s.Kind), opts.Type) {
			continue
		}

		var validationReasons []string
		if s.Status != claudefs.SkillStatusReady && len(s.ValidationReasons) > 0 {
			validationReasons = s.ValidationReasons
		}

		entries = append(entries, contract.SkillListEntry{
			Name:              s.Name,
			Type:              string(s.Kind),
			Scope:             string(s.Scope),
			Status:            string(s.Status),
			Project:           s.ProjectName,
			Path:              s.Path,
			ValidationReasons: validationReasons,
		})
	}
	return contract.SkillListResult{Skills: entries}, nil
}

// AgentsList returns the agents list projection.
func AgentsList(opts contract.AgentListOptions) (contract.AgentListResult, error) {
	agentResult := scanAgents()
	if agentResult.Err != nil {
		return contract.AgentListResult{}, statusError("scan agents: %v", agentResult.Err)
	}

	entries := make([]contract.AgentListEntry, 0, len(agentResult.Agents))
	for _, a := range agentResult.Agents {
		if opts.Project != "" && !strings.EqualFold(a.ProjectName, opts.Project) {
			continue
		}
		if opts.Scope != "" && !scopeMatches(string(a.Scope), opts.Scope) {
			continue
		}

		var validationReasons []string
		if a.Status != claudefs.AgentStatusReady && len(a.ValidationReasons) > 0 {
			validationReasons = a.ValidationReasons
		}

		entries = append(entries, contract.AgentListEntry{
			Name:              a.Name,
			Scope:             string(a.Scope),
			Status:            string(a.Status),
			Project:           a.ProjectName,
			Path:              a.Path,
			ValidationReasons: validationReasons,
		})
	}
	return contract.AgentListResult{Agents: entries}, nil
}

// AgentInspect returns the agent detail projection.
func AgentInspect(opts contract.AgentInspectOptions) (contract.AgentDetailResult, error) {
	agentResult := scanAgents()
	if agentResult.Err != nil {
		return contract.AgentDetailResult{}, statusError("scan agents: %v", agentResult.Err)
	}

	target := strings.ToLower(opts.Target)
	var found *claudefs.AgentResource
	for i := range agentResult.Agents {
		if strings.ToLower(agentResult.Agents[i].Name) == target || agentResult.Agents[i].Path == opts.Target {
			found = &agentResult.Agents[i]
			break
		}
	}
	if found == nil {
		return contract.AgentDetailResult{}, statusError("agent not found: %s", opts.Target)
	}

	return contract.AgentDetailResult{
		AgentDetail: contract.AgentDetail{
			Name:    found.Name,
			Scope:   string(found.Scope),
			Source:  string(found.Source),
			Status:  string(found.Status),
			Project: found.ProjectName,
			Path:    found.Path,
			Configuration: contract.AgentConfiguration{
				Model:      found.Model,
				Tools:      found.Tools,
				Permission: found.PermissionMode,
			},
			Validation: contract.AgentValidation{
				Valid:   found.Status == claudefs.AgentStatusReady,
				Reasons: found.ValidationReasons,
			},
		},
	}, nil
}

func scopeMatches(value, query string) bool {
	v := strings.ToLower(value)
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" || v == "" {
		return false
	}
	return strings.Contains(v, q) || strings.Contains(q, v)
}
