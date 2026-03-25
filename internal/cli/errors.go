package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

// CLIError is a command execution error that can render as text or JSON.
type CLIError struct {
	Message string
}

func (e CLIError) Error() string {
	return e.Message
}

func exitWithError(err error, jsonMode bool) {
	writeError(os.Stdout, os.Stderr, err, jsonMode)
	os.Exit(1)
}

func writeError(stdout, stderr *os.File, err error, jsonMode bool) {
	msg := err.Error()
	if cliErr, ok := err.(CLIError); ok {
		msg = cliErr.Message
	}
	if jsonMode {
		renderJSONError(stdout, msg)
	} else {
		fmt.Fprintf(stderr, "error: %s\n", msg)
	}
}

// renderJSON encodes a value as compact JSON to the writer.
func renderJSON(w *os.File, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	_, err = w.Write([]byte("\n"))
	return err
}

// renderJSONError writes a JSON error payload to the writer.
func renderJSONError(w *os.File, message string) {
	payload := map[string]string{"error": message}
	if err := renderJSON(w, payload); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", message)
	}
}

// renderResult dispatches to the appropriate text or JSON renderer.
func renderResult(w *os.File, mode OutputMode, result CommandResult) {
	if mode == OutputJSON {
		renderJSONMode(w, result)
	} else {
		renderTextMode(w, result)
	}
}
