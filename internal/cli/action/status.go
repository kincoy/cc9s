package action

import (
	"fmt"
	"sort"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/cli/contract"
)

// Status computes the aggregate environment overview.
func Status() (contract.StatusResult, error) {
	scanResult := scanProjects()
	if scanResult.Err != nil {
		return contract.StatusResult{}, statusError("scan projects: %v", scanResult.Err)
	}

	health, err := computeHealthMetrics()
	if err != nil {
		health = claudefs.HealthMetrics{}
	}

	skillResult := scanSkills()
	agentResult := scanAgents()
	totalResources := len(skillResult.Skills) + len(agentResult.Agents)

	allSessions := make([]claudefs.Session, 0)
	for _, project := range scanResult.Projects {
		sessions, err := loadProjectSessions(project.EncodedPath)
		if err != nil {
			continue
		}
		allSessions = append(allSessions, sessions...)
	}

	lifecycle := claudefs.SummarizeLifecycleSessions(allSessions)
	var totalSize int64
	for _, p := range scanResult.Projects {
		totalSize += p.TotalSize
	}

	var issues []contract.StatusIssue
	totalSessions := lifecycle.Active + lifecycle.Idle + lifecycle.Completed + lifecycle.Stale
	if lifecycle.Stale > 0 {
		pct := ""
		if totalSessions > 0 {
			pct = fmt.Sprintf("%.0f%%", float64(lifecycle.Stale)/float64(totalSessions)*100)
		}
		issues = append(issues, contract.StatusIssue{
			Type:       "stale sessions",
			Count:      lifecycle.Stale,
			Percentage: pct,
			Suggestion: "Run: cc9s sessions cleanup --dry-run",
		})
	}

	if skillResult.InvalidCount > 0 {
		issues = append(issues, contract.StatusIssue{
			Type:       "invalid skills",
			Count:      skillResult.InvalidCount,
			Suggestion: "Run: cc9s skills list --json",
		})
	}
	if agentResult.InvalidCount > 0 {
		issues = append(issues, contract.StatusIssue{
			Type:       "invalid agents",
			Count:      agentResult.InvalidCount,
			Suggestion: "Run: cc9s agents list --json",
		})
	}

	topProjects := make([]contract.TopProject, 0, 5)
	for _, p := range scanResult.Projects {
		topProjects = append(topProjects, contract.TopProject{
			Name:      p.Name,
			Sessions:  p.SessionCount,
			Active:    p.ActiveSessionCount,
			SizeBytes: p.TotalSize,
		})
	}
	sortTopProjects(topProjects)
	if len(topProjects) > 5 {
		topProjects = topProjects[:5]
	}

	return contract.StatusResult{
		Projects:       len(scanResult.Projects),
		Sessions:       totalSessions,
		Resources:      totalResources,
		TotalSizeBytes: totalSize,
		Lifecycle: contract.LifecycleSummary{
			Active:    lifecycle.Active,
			Idle:      lifecycle.Idle,
			Completed: lifecycle.Completed,
			Stale:     lifecycle.Stale,
		},
		Issues:      issues,
		TopProjects: topProjects,
		Health:      &health,
	}, nil
}

func sortTopProjects(entries []contract.TopProject) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Sessions != entries[j].Sessions {
			return entries[i].Sessions > entries[j].Sessions
		}
		if entries[i].Active != entries[j].Active {
			return entries[i].Active > entries[j].Active
		}
		return entries[i].Name < entries[j].Name
	})
}
