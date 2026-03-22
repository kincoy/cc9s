package claudefs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DeleteTarget describes a session to be deleted.
type DeleteTarget struct {
	SessionID      string // Session ID
	EncodedPath    string // Project encoded path
	IsActive       bool   // Whether the session is active
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

	// Delete active marker file (if it exists)
	activePath := filepath.Join(projectDir, "sessions", target.SessionID+".json")
	if err := os.Remove(activePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove %s: %w", activePath, err)
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
