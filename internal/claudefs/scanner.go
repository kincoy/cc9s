package claudefs

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ScanProjects scans the ~/.claude/ directory and builds a project list.
func ScanProjects() ScanResult {
	start := time.Now()
	result := ScanResult{
		Projects: make([]Project, 0),
	}

	projectsDir := ProjectsDir()
	activeMarkers := getActiveSessionMarkers()

	// Scan project directories
	projects, totalSessions, activeCount, err := scanProjectsDir(projectsDir, activeMarkers, start)
	if err != nil {
		result.Err = err
		return result
	}

	// Sort by most recently active time (descending)
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastActiveAt.After(projects[j].LastActiveAt)
	})

	result.Projects = projects
	result.TotalSessions = totalSessions
	result.ActiveCount = activeCount
	result.ScanDuration = time.Since(start)

	return result
}

// scanProjectsDir scans the projects directory.
func scanProjectsDir(projectsDir string, activeMarkers map[string]activeSessionInfo, now time.Time) ([]Project, int, int, error) {
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, 0, 0, err
	}

	projects := make([]Project, 0, len(entries))
	totalSessions := 0
	activeCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Decode directory name to get the project path
		encodedPath := entry.Name()
		decodedPath := DecodePathFS(encodedPath)
		if decodedPath == "" {
			// JSONL fallback: read cwd from a JSONL file in the project directory
			decodedPath = extractCwdFromProjectDir(filepath.Join(projectsDir, encodedPath))
		}
		projectName := ExtractProjectName(decodedPath)

		// Scan session files under this project
		projectDir := filepath.Join(projectsDir, encodedPath)
		sessionCount, totalSize, lastActive, projectActiveCount := scanProjectSessions(projectDir, activeMarkers, now)
		resourceSummary := summarizeProjectResources(decodedPath)

		if sessionCount > 0 {
			projects = append(projects, Project{
				Name:               projectName,
				Path:               decodedPath,
				EncodedPath:        encodedPath,
				SessionCount:       sessionCount,
				ActiveSessionCount: projectActiveCount,
				LastActiveAt:       lastActive,
				TotalSize:          totalSize,
				SkillCount:         resourceSummary.SkillCount,
				CommandCount:       resourceSummary.CommandCount,
				AgentCount:         resourceSummary.AgentCount,
				HasSkillsRoot:      resourceSummary.HasSkillsRoot,
				HasCommandsRoot:    resourceSummary.HasCommandsRoot,
				HasAgentsRoot:      resourceSummary.HasAgentsRoot,
			})
			totalSessions += sessionCount
			activeCount += projectActiveCount
		}
	}

	return projects, totalSessions, activeCount, nil
}

// scanProjectSessions scans session files under a single project.
func scanProjectSessions(projectDir string, activeMarkers map[string]activeSessionInfo, now time.Time) (count int, totalSize int64, lastActive time.Time, activeCount int) {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return 0, 0, time.Time{}, 0
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only count .jsonl files
		if !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		count++

		// Get file info
		filePath := filepath.Join(projectDir, entry.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		totalSize += info.Size()

		// Update most recent active time
		if info.ModTime().After(lastActive) {
			lastActive = info.ModTime()
		}

		sessionID := strings.TrimSuffix(entry.Name(), ".jsonl")
		snapshot := ClassifySessionLifecycle(
			info.ModTime(),
			hasActiveMarker(activeMarkers, sessionID),
			SessionHealth{IsReliable: true},
			now,
			DefaultActivityWindow,
		)
		if snapshot.State == SessionLifecycleActive {
			activeCount++
		}
	}

	return count, totalSize, lastActive, activeCount
}

type projectResourceSummary struct {
	SkillCount      int
	CommandCount    int
	AgentCount      int
	HasSkillsRoot   bool
	HasCommandsRoot bool
	HasAgentsRoot   bool
}

func summarizeProjectResources(projectPath string) projectResourceSummary {
	projectPath = strings.TrimSpace(projectPath)
	if projectPath == "" {
		return projectResourceSummary{}
	}

	skillsRoot := filepath.Join(projectPath, ".claude", "skills")
	commandsRoot := filepath.Join(projectPath, ".claude", "commands")
	agentsRoot := filepath.Join(projectPath, ".claude", "agents")

	skillCount, hasSkillsRoot := countProjectSkillResources(skillsRoot)
	commandCount, hasCommandsRoot := countMarkdownResources(commandsRoot)
	agentCount, hasAgentsRoot := countMarkdownResources(agentsRoot)

	return projectResourceSummary{
		SkillCount:      skillCount,
		CommandCount:    commandCount,
		AgentCount:      agentCount,
		HasSkillsRoot:   hasSkillsRoot,
		HasCommandsRoot: hasCommandsRoot,
		HasAgentsRoot:   hasAgentsRoot,
	}
}

func countProjectSkillResources(root string) (count int, exists bool) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0, false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			count++
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
			count++
		}
	}

	return count, true
}

func countMarkdownResources(root string) (count int, exists bool) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0, false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
			count++
		}
	}

	return count, true
}

// LoadProjectSessions loads all sessions under a project.
func LoadProjectSessions(projectEncodedPath string) ([]Session, error) {
	// Project directory path
	projectDir := filepath.Join(ProjectsDir(), projectEncodedPath)

	// Scan JSONL files
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return nil, err
	}

	// Build active session set
	activeMarkers := getActiveSessionMarkers()

	// Project-level path resolution (no longer reading JSONL per session)
	projectPath := DecodePathFS(projectEncodedPath)
	if projectPath == "" {
		projectPath = extractCwdFromProjectDir(projectDir)
	}

	sessions := make([]Session, 0)
	now := time.Now()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		// Session ID = filename without .jsonl extension
		sessionID := strings.TrimSuffix(entry.Name(), ".jsonl")
		filePath := filepath.Join(projectDir, entry.Name())

		// Get file info
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		inspection := inspectSessionFile(filePath)
		hasMarker := hasActiveMarker(activeMarkers, sessionID)
		lifecycle := ClassifySessionLifecycle(info.ModTime(), hasMarker, inspection.Health, now, DefaultActivityWindow)

		// Estimate event count (simplified: file size / average line size of 200 bytes)
		eventCount := int(info.Size() / 200)
		if eventCount < 1 && info.Size() > 0 {
			eventCount = 1
		}

		sessions = append(sessions, Session{
			ID:              sessionID,
			ProjectPath:     projectPath,
			EncodedPath:     projectEncodedPath,
			StartTime:       info.ModTime(), // Simplified: using ModTime, actual creation time requires platform-specific API
			LastActiveAt:    info.ModTime(),
			EventCount:      eventCount,
			FileSize:        info.Size(),
			HasActiveMarker: hasMarker,
			IsActive:        lifecycle.State == SessionLifecycleActive,
			Lifecycle:       lifecycle,
			Summary:         inspection.Summary,
		})
	}

	// Sort by most recently active time (descending)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActiveAt.After(sessions[j].LastActiveAt)
	})

	return sessions, nil
}

// activeSessionInfo is the JSON structure of an active session marker file.
type activeSessionInfo struct {
	PID        int    `json:"pid"`
	SessionID  string `json:"sessionId"`
	CWD        string `json:"cwd"`
	StartedAt  int64  `json:"startedAt"`
	MarkerPath string `json:"-"`
}

// getActiveSessionMarkers reads all active session marker files keyed by session ID.
func getActiveSessionMarkers() map[string]activeSessionInfo {
	activeMarkers := make(map[string]activeSessionInfo)
	sessionsDir := SessionsDir()

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return activeMarkers
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Skip non-PID files (e.g. compaction-log.txt)
		if entry.Name() == "compaction-log.txt" {
			continue
		}

		// Read and parse JSON
		filePath := filepath.Join(sessionsDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var info activeSessionInfo
		if err := json.Unmarshal(data, &info); err != nil {
			continue
		}

		info.MarkerPath = filePath

		// Record the newest marker for each session ID.
		if info.SessionID != "" {
			existing, ok := activeMarkers[info.SessionID]
			if !ok || info.StartedAt >= existing.StartedAt {
				activeMarkers[info.SessionID] = info
			}
		}
	}

	return activeMarkers
}

func hasActiveMarker(activeMarkers map[string]activeSessionInfo, sessionID string) bool {
	_, ok := activeMarkers[sessionID]
	return ok
}

// sessionMetadata is the session metadata from the first line of a JSONL file.
type sessionMetadata struct {
	SessionID string `json:"sessionId"`
	CWD       string `json:"cwd"`
	GitBranch string `json:"gitBranch"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

// extractCwdFromSession reads the cwd field from a JSONL file.
func extractCwdFromSession(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Only read the first line
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	if !scanner.Scan() {
		return ""
	}

	var meta sessionMetadata
	if err := json.Unmarshal(scanner.Bytes(), &meta); err != nil {
		return ""
	}

	return meta.CWD
}

// extractCwdFromProjectDir reads the cwd field from the first JSONL file in a project directory.
func extractCwdFromProjectDir(projectDir string) string {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		return extractCwdFromSession(filepath.Join(projectDir, entry.Name()))
	}

	return ""
}
