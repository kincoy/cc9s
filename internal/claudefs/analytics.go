package claudefs

import (
	"fmt"
	"sort"
	"time"
)

// ComputeHealthMetrics calculates environment-wide health insights.
func ComputeHealthMetrics(homeDir string) (HealthMetrics, error) {
	_ = homeDir

	scanResult := ScanProjects()
	if scanResult.Err != nil {
		return HealthMetrics{}, scanResult.Err
	}

	var globalSessions []GlobalSession
	for _, proj := range scanResult.Projects {
		sessions, err := LoadProjectSessions(proj.EncodedPath)
		if err != nil {
			continue
		}
		for _, s := range sessions {
			globalSessions = append(globalSessions, GlobalSession{
				Session:     s,
				ProjectName: proj.Name,
			})
		}
	}

	return computeHealthMetrics(scanResult.Projects, globalSessions), nil
}

func computeHealthMetrics(projects []Project, sessions []GlobalSession) HealthMetrics {
	sessionsByProject := groupSessionsByProject(sessions)

	var projectScores []ProjectHealthScore
	totalWeightedScore := 0.0
	totalWeight := 0

	for _, proj := range projects {
		projSessions := sessionsByProject[proj.Name]
		if len(projSessions) == 0 {
			continue
		}

		score := computeProjectHealthScore(proj, projSessions)
		projectScores = append(projectScores, score)

		totalWeightedScore += float64(score.HealthScore) * float64(score.SessionCount)
		totalWeight += score.SessionCount
	}

	sort.Slice(projectScores, func(i, j int) bool {
		return projectScores[i].HealthScore < projectScores[j].HealthScore
	})

	envScore := 0
	if totalWeight > 0 {
		envScore = int(totalWeightedScore / float64(totalWeight))
	}

	return HealthMetrics{
		EnvironmentScore: envScore,
		ProjectScores:    projectScores,
		Recommendations:  generateRecommendations(envScore, projectScores),
	}
}

func groupSessionsByProject(sessions []GlobalSession) map[string][]Session {
	m := make(map[string][]Session)
	for _, gs := range sessions {
		m[gs.ProjectName] = append(m[gs.ProjectName], gs.Session)
	}
	return m
}

func computeProjectHealthScore(proj Project, sessions []Session) ProjectHealthScore {
	lifecycle := LifecycleSummary{Total: len(sessions)}
	for _, s := range sessions {
		switch s.Lifecycle.State {
		case SessionLifecycleActive:
			lifecycle.Active++
		case SessionLifecycleIdle:
			lifecycle.Idle++
		case SessionLifecycleCompleted:
			lifecycle.Completed++
		case SessionLifecycleStale:
			lifecycle.Stale++
		}
	}

	keepCount := 0
	for _, s := range sessions {
		stats, err := ParseSessionStats(s)
		if err != nil {
			continue
		}
		assessment := AssessSessionValue(*stats, s.Lifecycle, s.FileSize)
		if assessment.Recommendation == RecommendKeep {
			keepCount++
		}
	}

	activityScore := computeProjectActivityScore(proj, sessions)
	staleRatio := float64(lifecycle.Stale) / float64(lifecycle.Total)
	keepRatio := float64(keepCount) / float64(lifecycle.Total)

	staleScore := (1 - staleRatio) * 100
	keepScore := keepRatio * 100
	healthScore := (staleScore * 0.4) + (keepScore * 0.35) + (float64(activityScore) * 0.25)

	return ProjectHealthScore{
		ProjectName:      proj.Name,
		HealthScore:      int(healthScore),
		StaleRatio:       staleRatio,
		KeepRatio:        keepRatio,
		ActivityScore:    activityScore,
		SessionCount:     lifecycle.Total,
		LifecycleSummary: lifecycle,
	}
}

func computeProjectActivityScore(proj Project, sessions []Session) int {
	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	fourteenDaysAgo := now.AddDate(0, 0, -14)

	sessionsLast7d := filterSessionsByTime(sessions, sevenDaysAgo, now)
	sessionsPrev7d := filterSessionsByTime(sessions, fourteenDaysAgo, sevenDaysAgo)

	activeCount := 0
	idleCount := 0
	for _, s := range sessions {
		if s.Lifecycle.State == SessionLifecycleActive {
			activeCount++
		} else if s.Lifecycle.State == SessionLifecycleIdle {
			idleCount++
		}
	}

	activeIdleRatio := 0.0
	if activeCount+idleCount > 0 {
		activeIdleRatio = float64(activeCount) / float64(activeCount+idleCount)
	}

	growthRate := 0.0
	if len(sessionsPrev7d) > 0 {
		growthRate = (float64(len(sessionsLast7d))/float64(len(sessionsPrev7d)) - 1) * 100
	}

	density := 0.0
	if len(sessions) > 0 {
		density = float64(len(sessionsLast7d)) / float64(len(sessions)) * 100
	}

	score := (density * 0.5) + (activeIdleRatio * 100 * 0.3) + (growthRate * 0.2)
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return int(score)
}

func filterSessionsByTime(sessions []Session, start, end time.Time) []Session {
	var result []Session
	for _, s := range sessions {
		if !s.StartTime.Before(start) && s.StartTime.Before(end) {
			result = append(result, s)
		}
	}
	return result
}

func generateRecommendations(envScore int, projectScores []ProjectHealthScore) []string {
	var recs []string

	if envScore < 50 {
		recs = append(recs, "Environment health is low; clean up stale sessions")
	} else if envScore < 70 {
		recs = append(recs, "Environment health is moderate; review low-value sessions")
	}

	for i := 0; i < 3 && i < len(projectScores); i++ {
		ps := projectScores[i]
		if ps.HealthScore < 60 {
			if ps.StaleRatio > 0.3 {
				recs = append(recs, fmt.Sprintf(
					"Project %s: %.0f%% of sessions are stale; cleanup recommended",
					ps.ProjectName, ps.StaleRatio*100))
			}
			if ps.ActivityScore < 30 {
				recs = append(recs, fmt.Sprintf(
					"Project %s: low activity score; archive review recommended",
					ps.ProjectName))
			}
		}
	}

	if len(recs) == 0 {
		recs = append(recs, "Environment health looks good")
	}

	return recs
}
