package action

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kincoy/cc9s/internal/claudefs"
)

var (
	scanProjects         = claudefs.ScanProjects
	loadProjectSessions  = claudefs.LoadProjectSessions
	computeHealthMetrics = func() (claudefs.HealthMetrics, error) { return claudefs.ComputeHealthMetrics() }
	scanSkills           = claudefs.ScanSkills
	scanAgents           = claudefs.ScanAgents
	parseSessionStats    = claudefs.ParseSessionStats
	quickAssessSession   = claudefs.QuickAssessSession
	nowFunc              = time.Now
	formatTimeRFC3339    = func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format(time.RFC3339)
	}
	parseOlderThanDuration = func(value string) (time.Duration, error) {
		if value == "" {
			return 0, nil
		}
		if strings.HasSuffix(value, "d") {
			days, err := strconv.Atoi(strings.TrimSuffix(value, "d"))
			if err != nil || days < 0 {
				return 0, fmt.Errorf("invalid duration: %s", value)
			}
			return time.Duration(days) * 24 * time.Hour, nil
		}
		return time.ParseDuration(value)
	}
)

func statusError(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
