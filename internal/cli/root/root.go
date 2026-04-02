package root

import (
	"io"

	"github.com/kincoy/cc9s/internal/cli/command"
	"github.com/spf13/cobra"
)

func New(stdout, stderr io.Writer) *cobra.Command {
	return command.New(stdout, stderr)
}
