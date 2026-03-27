package claudefs

import "fmt"

// QuickAssessSession provides a lightweight value assessment using only Session fields.
func QuickAssessSession(s Session) SessionAssessment {
	if s.Lifecycle.State == SessionLifecycleActive || s.Lifecycle.State == SessionLifecycleIdle {
		return SessionAssessment{
			Score:          100,
			Recommendation: RecommendKeep,
			Reasons:        []string{fmt.Sprintf("session is %s", s.Lifecycle.State)},
		}
	}

	if s.Lifecycle.State == SessionLifecycleStale {
		reasons := []string{"session data is unreliable (Stale)"}
		reasons = append(reasons, s.Lifecycle.Evidence.Reasons...)
		return SessionAssessment{
			Score:          0,
			Recommendation: RecommendDelete,
			Reasons:        reasons,
		}
	}

	score := 0
	var reasons []string

	switch {
	case s.EventCount >= 200:
		score += 35
	case s.EventCount >= 50:
		score += 25
	case s.EventCount >= 20:
		score += 15
	case s.EventCount >= 8:
		score += 8
	default:
		reasons = append(reasons, fmt.Sprintf("few events (%d)", s.EventCount))
	}

	switch {
	case s.FileSize >= 100000:
		score += 35
	case s.FileSize >= 30000:
		score += 25
	case s.FileSize >= 10000:
		score += 15
	case s.FileSize >= 3000:
		score += 5
	default:
		reasons = append(reasons, "small session file")
	}

	switch {
	case len(s.Summary) >= 60:
		score += 20
	case len(s.Summary) >= 30:
		score += 12
	case len(s.Summary) >= 10:
		score += 5
	default:
		reasons = append(reasons, "brief or missing summary")
	}

	if s.Summary == "" || s.Summary == "-" {
		reasons = append(reasons, "no summary")
	} else {
		score += 10
	}

	if score > 100 {
		score = 100
	}

	recommendation := scoreToRecommendation(score)
	switch recommendation {
	case RecommendKeep:
		reasons = append(reasons, fmt.Sprintf("substantial session (%d events, %s)", s.EventCount, FormatSize(s.FileSize)))
	case RecommendMaybe:
		reasons = append(reasons, "moderate content - review before deleting")
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
