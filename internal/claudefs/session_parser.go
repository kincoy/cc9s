package claudefs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type sessionInspection struct {
	Summary string
	Health  SessionHealth
}

// ParseSessionStats parses session statistics (single-pass JSONL traversal).
func ParseSessionStats(session Session) (*SessionStats, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get user home dir: %w", err)
	}

	// Find the JSONL file
	jsonlPath, err := findSessionJSONL(homeDir, session.ID)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(jsonlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open session file: %w", err)
	}
	defer file.Close()

	stats := &SessionStats{
		SessionID: session.ID,
		Lifecycle: session.Lifecycle,
		ToolUsage: make(map[string]int),
	}

	scanner := bufio.NewScanner(file)
	firstTimestamp := true

	for scanner.Scan() {
		var event map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			// Skip corrupted lines
			continue
		}

		eventType, _ := event["type"].(string)

		// Extract metadata from any event that has version/gitBranch (usually in the first few lines)
		if stats.Version == "" {
			if v, ok := event["version"].(string); ok {
				stats.Version = v
			}
		}
		if stats.GitBranch == "" {
			if b, ok := event["gitBranch"].(string); ok {
				stats.GitBranch = b
			}
		}

		// Extract the first valid timestamp as the start time
		if firstTimestamp {
			if ts, ok := event["timestamp"].(string); ok {
				if t, err := time.Parse(time.RFC3339, ts); err == nil {
					stats.StartTime = t
					firstTimestamp = false
				}
			}
		}

		// Update last active time
		if ts, ok := event["timestamp"].(string); ok {
			if t, err := time.Parse(time.RFC3339, ts); err == nil {
				if t.After(stats.LastActiveTime) {
					stats.LastActiveTime = t
				}
			}
		}

		// Statistics by event type
		switch eventType {
		case "custom-title":
			if title, ok := event["customTitle"].(string); ok {
				stats.CustomTitle = title
			}

		case "assistant":
			// Extract model and token statistics
			if msg, ok := event["message"].(map[string]interface{}); ok {
				if model, ok := msg["model"].(string); ok && stats.Model == "" {
					stats.Model = model
				}

				// Token statistics
				if usage, ok := msg["usage"].(map[string]interface{}); ok {
					if input, ok := usage["input_tokens"].(float64); ok {
						stats.InputTokens += int(input)
					}
					if output, ok := usage["output_tokens"].(float64); ok {
						stats.OutputTokens += int(output)
					}
					if cache, ok := usage["cache_creation_input_tokens"].(float64); ok {
						stats.CacheTokens += int(cache)
					}
					if cacheRead, ok := usage["cache_read_input_tokens"].(float64); ok {
						stats.CacheTokens += int(cacheRead)
					}
				}

				// Tool call statistics
				if content, ok := msg["content"].([]interface{}); ok {
					for _, block := range content {
						if blockMap, ok := block.(map[string]interface{}); ok {
							if blockType, ok := blockMap["type"].(string); ok && blockType == "tool_use" {
								stats.ToolCallCount++
								if name, ok := blockMap["name"].(string); ok {
									stats.ToolUsage[name]++
								}
							}
						}
					}
				}
			}

		case "user":
			// Count user messages (excluding tool_result and isMeta)
			isMeta, _ := event["isMeta"].(bool)
			if !isMeta {
				if msg, ok := event["message"].(map[string]interface{}); ok {
					if content, ok := msg["content"].([]interface{}); ok {
						// Check if it is a tool_result
						isToolResult := false
						for _, block := range content {
							if blockMap, ok := block.(map[string]interface{}); ok {
								if blockType, ok := blockMap["type"].(string); ok && blockType == "tool_result" {
									isToolResult = true
									break
								}
							}
						}
						if !isToolResult {
							stats.TurnCount++
							stats.UserMsgCount++
						}
					} else if _, ok := msg["content"].(string); ok {
						// Plain text message
						stats.TurnCount++
						stats.UserMsgCount++
					}
				}
			}

		case "system":
			subtype, _ := event["subtype"].(string)
			switch subtype {
			case "compact_boundary":
				stats.CompactCount++
			case "api_error":
				stats.ErrorCount++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading session file: %w", err)
	}

	// Calculate duration
	if !stats.StartTime.IsZero() && !stats.LastActiveTime.IsZero() {
		stats.Duration = stats.LastActiveTime.Sub(stats.StartTime)
	}

	return stats, nil
}

// findSessionJSONL locates the JSONL file for a session.
func findSessionJSONL(homeDir, sessionID string) (string, error) {
	projectsDir := filepath.Join(homeDir, ".claude", "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read projects directory: %w", err)
	}

	// Iterate over all project directories
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		jsonlPath := filepath.Join(projectsDir, entry.Name(), sessionID+".jsonl")
		if _, err := os.Stat(jsonlPath); err == nil {
			return jsonlPath, nil
		}
	}

	return "", fmt.Errorf("session JSONL not found for session ID: %s", sessionID)
}

// ParseSessionLog parses session log organized by turns.
// offset: starting turn number (for pagination)
// limit: number of turns to return
// Returns: log entries, total turn count, error
func ParseSessionLog(sessionID string, offset, limit int) ([]LogEntry, int, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, 0, err
	}

	jsonlPath, err := findSessionJSONL(homeDir, sessionID)
	if err != nil {
		return nil, 0, err
	}

	file, err := os.Open(jsonlPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open session file: %w", err)
	}
	defer file.Close()

	// First pass scan: organize turns
	var allTurns []LogEntry
	currentTurn := &LogEntry{TurnNumber: 0}
	turnStarted := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}

		eventType, _ := event["type"].(string)

		// Filter out noise events
		if eventType == "file-history-snapshot" {
			continue
		}
		if eventType == "progress" {
			if data, ok := event["data"].(map[string]interface{}); ok {
				if dataType, ok := data["type"].(string); ok && dataType == "hook_progress" {
					continue
				}
			}
		}

		// User message: start a new turn
		if eventType == "user" {
			isMeta, _ := event["isMeta"].(bool)
			if !isMeta {
				// Check if it is a tool_result (not a new turn)
				isToolResult := false
				if msg, ok := event["message"].(map[string]interface{}); ok {
					if content, ok := msg["content"].([]interface{}); ok {
						for _, block := range content {
							if blockMap, ok := block.(map[string]interface{}); ok {
								if blockType, ok := blockMap["type"].(string); ok && blockType == "tool_result" {
									isToolResult = true
									break
								}
							}
						}
					}
				}

				if !isToolResult {
					// Save the previous turn
					if turnStarted {
						allTurns = append(allTurns, *currentTurn)
					}

					// Start a new turn
					currentTurn = &LogEntry{
						TurnNumber: len(allTurns) + 1,
						ToolCalls:  []ToolCall{},
					}
					turnStarted = true

					// Extract timestamp
					if ts, ok := event["timestamp"].(string); ok {
						currentTurn.Timestamp, _ = time.Parse(time.RFC3339, ts)
					}

					// Extract user message
					if msg, ok := event["message"].(map[string]interface{}); ok {
						if content, ok := msg["content"].(string); ok {
							currentTurn.UserMsg = truncateString(content, 200)
						} else if content, ok := msg["content"].([]interface{}); ok {
							if len(content) > 0 {
								if block, ok := content[0].(map[string]interface{}); ok {
									if text, ok := block["text"].(string); ok {
										currentTurn.UserMsg = truncateString(text, 200)
									}
								}
							}
						}
					}
				}
			}
		}

		// Assistant message
		if eventType == "assistant" && turnStarted {
			if msg, ok := event["message"].(map[string]interface{}); ok {
				// Extract tool_use
				if content, ok := msg["content"].([]interface{}); ok {
					for _, block := range content {
						if blockMap, ok := block.(map[string]interface{}); ok {
							blockType, _ := blockMap["type"].(string)
							switch blockType {
							case "text":
								if text, ok := blockMap["text"].(string); ok {
									if currentTurn.AssistantMsg == "" {
										currentTurn.AssistantMsg = truncateString(text, 200)
									}
								}
							case "tool_use":
								toolCall := ToolCall{}
								if name, ok := blockMap["name"].(string); ok {
									toolCall.Name = name
								}
								if input, ok := blockMap["input"].(map[string]interface{}); ok {
									inputJSON, _ := json.Marshal(input)
									toolCall.Input = truncateString(string(inputJSON), 100)
								}
								currentTurn.ToolCalls = append(currentTurn.ToolCalls, toolCall)
							}
						}
					}
				}
			}
		}

		// system/turn_duration
		if eventType == "system" && turnStarted {
			if subtype, ok := event["subtype"].(string); ok && subtype == "turn_duration" {
				if duration, ok := event["durationMs"].(float64); ok {
					currentTurn.Duration = int(duration)
				}
			}
		}
	}

	// Save the last turn
	if turnStarted {
		allTurns = append(allTurns, *currentTurn)
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, fmt.Errorf("error reading session file: %w", err)
	}

	totalTurns := len(allTurns)

	// Pagination
	if offset >= totalTurns {
		return []LogEntry{}, totalTurns, nil
	}

	end := offset + limit
	if end > totalTurns {
		end = totalTurns
	}

	return allTurns[offset:end], totalTurns, nil
}

func inspectSessionFile(filePath string) sessionInspection {
	inspection := sessionInspection{
		Health: SessionHealth{
			IsReliable: false,
			Problem:    "The session file could not be inspected.",
		},
	}

	file, err := os.Open(filePath)
	if err != nil {
		inspection.Health.Problem = "The session file cannot be opened."
		return inspection
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	totalLines := 0
	parsedLines := 0
	userEvents := 0
	assistantEvents := 0
	hasSessionMeta := false

	for scanner.Scan() {
		totalLines++

		var event map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		parsedLines++

		if v, ok := event["sessionId"].(string); ok && v != "" {
			hasSessionMeta = true
		}
		if v, ok := event["cwd"].(string); ok && v != "" {
			hasSessionMeta = true
		}
		if v, ok := event["version"].(string); ok && v != "" {
			hasSessionMeta = true
		}

		eventType, _ := event["type"].(string)
		switch eventType {
		case "user":
			isMeta, _ := event["isMeta"].(bool)
			if !isMeta {
				userEvents++
				if inspection.Summary == "" {
					if summary := extractSummaryText(event); summary != "" {
						inspection.Summary = summary
					}
				}
			}
		case "assistant":
			assistantEvents++
		}

		if totalLines >= 200 && hasSessionMeta && (userEvents > 0 || assistantEvents > 0) {
			break
		}
	}

	switch {
	case totalLines == 0:
		inspection.Health.Problem = "The session file is empty."
	case parsedLines == 0:
		inspection.Health.Problem = "The session file has no parseable JSONL events."
	case userEvents == 0 && assistantEvents == 0:
		inspection.Health.Problem = "The session file only has residue events and no normal user/assistant session chain."
	case !hasSessionMeta:
		inspection.Health.Problem = "The session file is missing basic session metadata."
	default:
		inspection.Health.IsReliable = true
		inspection.Health.Problem = ""
	}

	return inspection
}

// ExtractSessionSummary extracts the first user message from a JSONL file as the session summary.
func ExtractSessionSummary(filePath string) string {
	return inspectSessionFile(filePath).Summary
}

// cleanSummary cleans summary text: removes newlines and extra spaces (does not truncate, UI layer handles that)
func cleanSummary(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// truncateString truncates a string to the specified number of runes (UTF-8 safe).
func truncateString(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}

func extractSummaryText(event map[string]interface{}) string {
	msg, ok := event["message"].(map[string]interface{})
	if !ok {
		return ""
	}

	var text string
	switch content := msg["content"].(type) {
	case string:
		text = content
	case []interface{}:
		for _, block := range content {
			if blockMap, ok := block.(map[string]interface{}); ok {
				if blockType, _ := blockMap["type"].(string); blockType == "text" {
					if t, ok := blockMap["text"].(string); ok {
						text = t
					}
				}
			}
			if text != "" {
				break
			}
		}
	}

	if text == "" {
		return ""
	}

	cleaned := cleanSummary(text)
	if cleaned == "" {
		return ""
	}
	if strings.HasPrefix(text, "<command-name>/") {
		return ""
	}
	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, "@") && !strings.Contains(trimmed, " ") {
		return ""
	}

	return cleaned
}
