package action

import (
	"sort"
	"strings"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/cli/contract"
)

// ProjectsList returns the project list projection.
func ProjectsList(opts contract.ProjectListOptions) (contract.ProjectListResult, error) {
	scanResult := scanProjects()
	if scanResult.Err != nil {
		return contract.ProjectListResult{}, statusError("scan projects: %v", scanResult.Err)
	}

	entries := make([]contract.ProjectListEntry, 0, len(scanResult.Projects))
	for _, p := range scanResult.Projects {
		entries = append(entries, contract.ProjectListEntry{
			Name:               p.Name,
			SessionCount:       p.SessionCount,
			ActiveSessionCount: p.ActiveSessionCount,
			LastActiveAt:       formatTimeRFC3339(p.LastActiveAt),
			SkillCount:         p.SkillCount,
			CommandCount:       p.CommandCount,
			AgentCount:         p.AgentCount,
			TotalSizeBytes:     p.TotalSize,
			Path:               p.Path,
		})
	}

	if opts.Sort != "" {
		entries = sortProjectEntries(entries, opts.Sort)
	}
	if opts.Limit > 0 && opts.Limit < len(entries) {
		entries = entries[:opts.Limit]
	}

	return contract.ProjectListResult{Projects: entries}, nil
}

// ProjectInspect returns the project detail projection.
func ProjectInspect(opts contract.ProjectInspectOptions) (contract.ProjectDetailResult, error) {
	scanResult := scanProjects()
	if scanResult.Err != nil {
		return contract.ProjectDetailResult{}, statusError("scan projects: %v", scanResult.Err)
	}

	target := strings.ToLower(opts.Target)
	var found *claudefs.Project
	for i := range scanResult.Projects {
		if strings.ToLower(scanResult.Projects[i].Name) == target || scanResult.Projects[i].Path == opts.Target {
			found = &scanResult.Projects[i]
			break
		}
	}
	if found == nil {
		return contract.ProjectDetailResult{}, statusError("project not found: %s", opts.Target)
	}

	sessions, err := loadProjectSessions(found.EncodedPath)
	if err != nil {
		return contract.ProjectDetailResult{}, statusError("load sessions: %v", err)
	}

	lifecycle := claudefs.SummarizeLifecycleSessions(sessions)
	return contract.ProjectDetailResult{
		ProjectDetail: contract.ProjectDetail{
			Name:           found.Name,
			Path:           found.Path,
			ClaudeRoot:     found.Path + "/.claude",
			LastActiveAt:   formatTimeRFC3339(found.LastActiveAt),
			TotalSizeBytes: found.TotalSize,
			Sessions: contract.ProjectSessions{
				Total:  found.SessionCount,
				Active: lifecycle.Active,
			},
			Resources: contract.ProjectResources{
				Skills:   found.SkillCount,
				Commands: found.CommandCount,
				Agents:   found.AgentCount,
			},
		},
	}, nil
}

func sortProjectEntries(entries []contract.ProjectListEntry, field string) []contract.ProjectListEntry {
	sorted := make([]contract.ProjectListEntry, len(entries))
	copy(sorted, entries)

	switch field {
	case "name":
		sort.Slice(sorted, func(i, j int) bool {
			return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
		})
	case "sessions":
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].SessionCount == sorted[j].SessionCount {
				return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
			}
			return sorted[i].SessionCount > sorted[j].SessionCount
		})
	}

	return sorted
}
