package render

import (
	"fmt"
	"io"

	"github.com/kincoy/cc9s/internal/cli/contract"
)

// ErrMessage normalizes an error into the message used for CLI output.
func ErrMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// WriteError renders an error to stdout in JSON mode or stderr in text mode.
func WriteError(stdout, stderr io.Writer, mode contract.OutputMode, err error) error {
	message := ErrMessage(err)
	if mode == contract.OutputJSON {
		return writeJSON(stdout, contract.ErrorPayload{Error: message})
	}

	if stderr == nil {
		return fmt.Errorf("nil stderr writer")
	}
	_, writeErr := fmt.Fprintf(stderr, "error: %s\n", message)
	return writeErr
}
