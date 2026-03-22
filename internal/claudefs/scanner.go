package claudefs

import (
	"bufio"
	"encoding/json"
	"fmt"
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

	// Get ~/.claude path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		result.Err = err
		return result
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	projectsDir := filepath.Join(claudeDir, "projects")

	// Scan project directories
	projects, totalSessions, err := scanProjectsDir(projectsDir)
	if err != nil {
		result.Err = err
		return result
	}

	// Count active sessions
	sessionsDir := filepath.Join(claudeDir, "sessions")
	activeCount := countActiveSessions(sessionsDir)

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
func scanProjectsDir(projectsDir string) ([]Project, int, error) {
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, 0, err
	}

	projects := make([]Project, 0, len(entries))
	totalSessions := 0

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
		sessionCount, totalSize, lastActive := scanProjectSessions(projectDir)

		if sessionCount > 0 {
			projects = append(projects, Project{
				Name:         projectName,
				Path:         decodedPath,
				EncodedPath:  encodedPath,
				SessionCount: sessionCount,
				LastActiveAt: lastActive,
				TotalSize:    totalSize,
			})
			totalSessions += sessionCount
		}
	}

	return projects, totalSessions, nil
}

// scanProjectSessions scans session files under a single project.
func scanProjectSessions(projectDir string) (count int, totalSize int64, lastActive time.Time) {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return 0, 0, time.Time{}
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
	}

	return count, totalSize, lastActive
}

// countActiveSessions counts the number of active sessions.
func countActiveSessions(sessionsDir string) int {
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			count++
		}
	}

	return count
}

// LoadProjectSessions loads all sessions under a project.
func LoadProjectSessions(projectEncodedPath string) ([]Session, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get user home dir: %w", err)
	}

	// Project directory path
	projectDir := filepath.Join(homeDir, ".claude", "projects", projectEncodedPath)

	// Scan JSONL files
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return nil, err
	}

	// Build active session set
	activeSessionIDs := getActiveSessionIDs(homeDir)

	// Project-level path resolution (no longer reading JSONL per session)
	projectPath := DecodePathFS(projectEncodedPath)
	if projectPath == "" {
		projectPath = extractCwdFromProjectDir(projectDir)
	}

	sessions := make([]Session, 0)

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

		// Check if the session is active
		isActive := activeSessionIDs[sessionID]

		// Estimate event count (simplified: file size / average line size of 200 bytes)
		eventCount := int(info.Size() / 200)
		if eventCount < 1 && info.Size() > 0 {
			eventCount = 1
		}

		// Extract session summary (first user message)
		summary := ExtractSessionSummary(filePath)

		sessions = append(sessions, Session{
			ID:           sessionID,
			ProjectPath:  projectPath,
			EncodedPath:  projectEncodedPath,
			StartTime:    info.ModTime(), // Simplified: using ModTime, actual creation time requires platform-specific API
			LastActiveAt: info.ModTime(),
			EventCount:   eventCount,
			FileSize:     info.Size(),
			IsActive:     isActive,
			Summary:      summary,
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
	PID       int    `json:"pid"`
	SessionID string `json:"sessionId"`
	CWD       string `json:"cwd"`
	StartedAt int64  `json:"startedAt"`
}

// getActiveSessionIDs reads all active session IDs.
func getActiveSessionIDs(homeDir string) map[string]bool {
	activeIDs := make(map[string]bool)
	sessionsDir := filepath.Join(homeDir, ".claude", "sessions")

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return activeIDs
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

		// Record the active session ID
		if info.SessionID != "" {
			activeIDs[info.SessionID] = true
		}
	}

	return activeIDs
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
