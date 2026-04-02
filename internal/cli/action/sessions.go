package action

import (
	"sort"
	"strings"
	"time"

	"github.com/kincoy/cc9s/internal/claudefs"
	"github.com/kincoy/cc9s/internal/cli/contract"
)

// SessionsList returns the sessions list projection.
func SessionsList(opts contract.SessionListOptions) (contract.SessionListResult, error) {
	scanResult := scanProjects()
	if scanResult.Err != nil {
		return contract.SessionListResult{}, statusError("scan projects: %v", scanResult.Err)
	}

	var allSessions []contract.SessionListEntry
	for _, project := range scanResult.Projects {
		if opts.Project != "" && !strings.EqualFold(project.Name, opts.Project) {
			continue
		}

		sessions, err := loadProjectSessions(project.EncodedPath)
		if err != nil {
			continue
		}

		for _, s := range sessions {
			if opts.State != "" && !claudefs.LifecycleStateMatchesQuery(s.Lifecycle.State, opts.State) {
				continue
			}

			summary := s.Summary
			if summary != "" {
				summary = truncateText(summary, 80)
			}

			allSessions = append(allSessions, contract.SessionListEntry{
				ID:           s.ID,
				Project:      project.Name,
				State:        string(s.Lifecycle.State),
				LastActiveAt: formatTimeRFC3339(s.LastActiveAt),
				Summary:      summary,
			})
		}
	}

	if opts.Sort != "" {
		allSessions = sortSessionEntries(allSessions, opts.Sort)
	}
	if opts.Limit > 0 && opts.Limit < len(allSessions) {
		allSessions = allSessions[:opts.Limit]
	}

	return contract.SessionListResult{Sessions: allSessions}, nil
}

// SessionInspect returns the session detail projection.
func SessionInspect(opts contract.SessionInspectOptions) (contract.SessionDetailResult, error) {
	scanResult := scanProjects()
	if scanResult.Err != nil {
		return contract.SessionDetailResult{}, statusError("scan projects: %v", scanResult.Err)
	}

	var found *claudefs.Session
	var projectName string
	for _, project := range scanResult.Projects {
		sessions, err := loadProjectSessions(project.EncodedPath)
		if err != nil {
			continue
		}
		for i := range sessions {
			if sessions[i].ID == opts.ID {
				found = &sessions[i]
				projectName = project.Name
				break
			}
		}
		if found != nil {
			break
		}
	}
	if found == nil {
		return contract.SessionDetailResult{}, statusError("session not found: %s", opts.ID)
	}

	stats, err := parseSessionStats(*found)
	if err != nil {
		return contract.SessionDetailResult{}, statusError("parse session stats: %v", err)
	}

	reasons := make([]string, len(found.Lifecycle.Evidence.Reasons))
	copy(reasons, found.Lifecycle.Evidence.Reasons)

	var durationSeconds float64
	if stats.Duration > 0 {
		durationSeconds = stats.Duration.Seconds()
	}

	return contract.SessionDetailResult{
		SessionDetail: contract.SessionDetail{
			ID:      found.ID,
			Project: projectName,
			Path:    found.ProjectPath,
			Summary: found.Summary,
			Lifecycle: contract.SessionLifecycleDetail{
				State:           string(found.Lifecycle.State),
				LastActiveAt:    formatTimeRFC3339(found.Lifecycle.Evidence.LastActiveAt),
				HasActiveMarker: found.HasActiveMarker,
				Reasons:         reasons,
			},
			Metadata: contract.SessionMetadata{
				Model:           stats.Model,
				Version:         stats.Version,
				GitBranch:       stats.GitBranch,
				CreatedAt:       formatTimeRFC3339(stats.StartTime),
				UpdatedAt:       formatTimeRFC3339(stats.LastActiveTime),
				DurationSeconds: durationSeconds,
			},
			Activity: contract.SessionActivity{
				TurnCount:        stats.TurnCount,
				UserMessageCount: stats.UserMsgCount,
				ToolCallCount:    stats.ToolCallCount,
				ToolUsage:        stats.ToolUsage,
			},
			Tokens: contract.SessionTokens{
				Input:  stats.InputTokens,
				Output: stats.OutputTokens,
				Cache:  stats.CacheTokens,
			},
		},
	}, nil
}

// SessionsCleanup returns the cleanup preview projection.
func SessionsCleanup(opts contract.CleanupOptions) (contract.CleanupResult, error) {
	scanResult := scanProjects()
	if scanResult.Err != nil {
		return contract.CleanupResult{}, statusError("scan projects: %v", scanResult.Err)
	}

	now := nowFunc()
	stateFilter := opts.State
	if stateFilter == "" {
		stateFilter = "stale"
	}

	var matches []contract.CleanupSessionMatch
	projectGroups := make(map[string]int)
	var totalSize int64

	for _, project := range scanResult.Projects {
		if opts.Project != "" && !strings.EqualFold(project.Name, opts.Project) {
			continue
		}

		sessions, err := loadProjectSessions(project.EncodedPath)
		if err != nil {
			continue
		}

		projectMatchCount := 0
		for _, s := range sessions {
			if !claudefs.LifecycleStateMatchesQuery(s.Lifecycle.State, stateFilter) {
				continue
			}

			if opts.OlderThan != "" {
				maxAge, err := parseOlderThanDuration(opts.OlderThan)
				if err != nil {
					return contract.CleanupResult{}, statusError("parse older-than: %v", err)
				}
				if now.Sub(s.LastActiveAt) < maxAge {
					continue
				}
			}

			assessment := quickAssessSession(s)
			ageHours := now.Sub(s.LastActiveAt).Hours()
			matches = append(matches, contract.CleanupSessionMatch{
				ID:             s.ID,
				Project:        project.Name,
				State:          string(s.Lifecycle.State),
				AgeHours:       ageHours,
				UpdatedAt:      formatTimeRFC3339(s.LastActiveAt),
				Recommendation: string(assessment.Recommendation),
				Score:          assessment.Score,
				Reasons:        assessment.Reasons,
			})
			totalSize += s.FileSize
			projectMatchCount++
		}

		if projectMatchCount > 0 {
			projectGroups[project.Name] = projectMatchCount
		}
	}

	projects := make([]contract.CleanupProjectGroup, 0, len(projectGroups))
	for name, count := range projectGroups {
		projects = append(projects, contract.CleanupProjectGroup{
			Name:         name,
			SessionCount: count,
		})
	}

	var deleteCount, maybeCount, keepCount int
	for _, match := range matches {
		switch claudefs.CleanupRecommendation(match.Recommendation) {
		case claudefs.RecommendDelete:
			deleteCount++
		case claudefs.RecommendMaybe:
			maybeCount++
		case claudefs.RecommendKeep:
			keepCount++
		}
	}

	return contract.CleanupResult{
		DryRun: true,
		Filters: contract.CleanupFilters{
			State:     stateFilter,
			OlderThan: opts.OlderThan,
			Project:   opts.Project,
		},
		Summary: contract.CleanupSummary{
			MatchedSessions: len(matches),
			MatchedProjects: len(projects),
			TotalSizeBytes:  totalSize,
			DeleteCount:     deleteCount,
			MaybeCount:      maybeCount,
			KeepCount:       keepCount,
		},
		Projects: projects,
		Sessions: matches,
	}, nil
}

func sortSessionEntries(entries []contract.SessionListEntry, field string) []contract.SessionListEntry {
	sorted := make([]contract.SessionListEntry, len(entries))
	copy(sorted, entries)

	switch field {
	case "updated":
		sort.Slice(sorted, func(i, j int) bool {
			ti, ei := time.Parse(time.RFC3339, sorted[i].LastActiveAt)
			tj, ej := time.Parse(time.RFC3339, sorted[j].LastActiveAt)
			if ei != nil && ej != nil {
				return false
			}
			if ei != nil {
				return false
			}
			if ej != nil {
				return true
			}
			return ti.After(tj)
		})
	case "state":
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].State == sorted[j].State {
				return sorted[i].LastActiveAt > sorted[j].LastActiveAt
			}
			return strings.ToLower(sorted[i].State) < strings.ToLower(sorted[j].State)
		})
	case "project":
		sort.Slice(sorted, func(i, j int) bool {
			if strings.ToLower(sorted[i].Project) == strings.ToLower(sorted[j].Project) {
				return sorted[i].LastActiveAt > sorted[j].LastActiveAt
			}
			return strings.ToLower(sorted[i].Project) < strings.ToLower(sorted[j].Project)
		})
	}

	return sorted
}

func truncateText(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}
