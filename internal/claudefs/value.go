package claudefs

import "fmt"

// AssessSessionValue scores a session's content value and returns a cleanup recommendation.
// Higher scores indicate more valuable sessions that should be kept.
func AssessSessionValue(stats SessionStats, lifecycle SessionLifecycleSnapshot, fileSize int64) SessionAssessment {
	if lifecycle.State == SessionLifecycleActive || lifecycle.State == SessionLifecycleIdle {
		return SessionAssessment{
			Score:          100,
			Recommendation: RecommendKeep,
			Reasons:        []string{fmt.Sprintf("session is %s", lifecycle.State)},
		}
	}

	if lifecycle.State == SessionLifecycleStale {
		reasons := []string{"session data is unreliable (Stale)"}
		reasons = append(reasons, lifecycle.Evidence.Reasons...)
		return SessionAssessment{
			Score:          0,
			Recommendation: RecommendDelete,
			Reasons:        reasons,
		}
	}

	score := 0
	var reasons []string

	switch {
	case stats.TurnCount >= 20:
		score += 30
	case stats.TurnCount >= 10:
		score += 20
	case stats.TurnCount >= 5:
		score += 12
	case stats.TurnCount >= 3:
		score += 5
	default:
		reasons = append(reasons, fmt.Sprintf("shallow conversation (%d turns)", stats.TurnCount))
	}

	switch {
	case stats.ToolCallCount >= 30:
		score += 30
	case stats.ToolCallCount >= 15:
		score += 22
	case stats.ToolCallCount >= 5:
		score += 12
	case stats.ToolCallCount >= 1:
		score += 5
	default:
		reasons = append(reasons, "no tool calls")
	}

	totalTokens := stats.InputTokens + stats.OutputTokens
	switch {
	case totalTokens >= 50000:
		score += 25
	case totalTokens >= 20000:
		score += 18
	case totalTokens >= 5000:
		score += 10
	case totalTokens >= 1000:
		score += 4
	default:
		reasons = append(reasons, "minimal token usage")
	}

	switch {
	case fileSize >= 100000:
		score += 15
	case fileSize >= 50000:
		score += 10
	case fileSize >= 10000:
		score += 5
	}

	if score > 100 {
		score = 100
	}

	recommendation := scoreToRecommendation(score)
	switch recommendation {
	case RecommendKeep:
		reasons = append(reasons, fmt.Sprintf("substantial session (%d turns, %d tool calls)", stats.TurnCount, stats.ToolCallCount))
	case RecommendMaybe:
		reasons = append(reasons, "moderate content value - review before deleting")
	}

	if len(reasons) == 0 {
		reasons = []string{"standard content value"}
	}

	return SessionAssessment{
		Score:          score,
		Recommendation: recommendation,
		Reasons:        reasons,
	}
}

func scoreToRecommendation(score int) CleanupRecommendation {
	switch {
	case score >= 60:
		return RecommendKeep
	case score >= 30:
		return RecommendMaybe
	default:
		return RecommendDelete
	}
}
