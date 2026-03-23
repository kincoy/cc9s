package claudefs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DeleteTarget describes a session to be deleted.
type DeleteTarget struct {
	SessionID       string // Session ID
	EncodedPath     string // Project encoded path
	HasActiveMarker bool   // Whether a raw marker file exists
	IsActive        bool   // Whether the session is currently active under lifecycle rules
}

// DeleteSession deletes a single session (JSONL file + active marker)
func DeleteSession(target DeleteTarget) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	projectDir := filepath.Join(homeDir, ".claude", "projects", target.EncodedPath)

	// Validate path does not escape the allowed directory
	projectDir = filepath.Clean(projectDir)
	allowedDir := filepath.Clean(filepath.Join(homeDir, ".claude", "projects"))
	if !strings.HasPrefix(projectDir, allowedDir+string(filepath.Separator)) && projectDir != allowedDir {
		return fmt.Errorf("invalid project path: %s", target.EncodedPath)
	}

	// Delete JSONL file
	jsonlPath := filepath.Join(projectDir, target.SessionID+".jsonl")
	if err := os.Remove(jsonlPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove %s: %w", jsonlPath, err)
	}

	// Delete matching marker files (including stale markers) when present.
	if target.HasActiveMarker {
		markers := getActiveSessionMarkers(homeDir)
		if marker, ok := markers[target.SessionID]; ok && marker.MarkerPath != "" {
			if err := os.Remove(marker.MarkerPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove %s: %w", marker.MarkerPath, err)
			}
		}
	}

	return nil
}

// DeleteSessions deletes multiple sessions in batch
func DeleteSessions(targets []DeleteTarget) (deleted int, errs []error) {
	for _, target := range targets {
		if err := DeleteSession(target); err != nil {
			shortID := target.SessionID
			if len(shortID) > 8 {
				shortID = shortID[:8]
			}
			errs = append(errs, fmt.Errorf("session %s: %w", shortID, err))
		} else {
			deleted++
		}
	}
	return deleted, errs
}
