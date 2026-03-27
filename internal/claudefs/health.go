package claudefs

// HealthMetrics captures environment and project-level health insights.
type HealthMetrics struct {
	EnvironmentScore int
	ProjectScores    []ProjectHealthScore
	Recommendations  []string
}

type ProjectHealthScore struct {
	ProjectName      string
	HealthScore      int
	StaleRatio       float64
	KeepRatio        float64
	ActivityScore    int
	SessionCount     int
	LifecycleSummary LifecycleSummary
}
