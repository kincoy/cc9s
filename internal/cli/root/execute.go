package root

import (
	"context"
	"io"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/kincoy/cc9s/internal/cli/contract"
	"github.com/kincoy/cc9s/internal/cli/render"
	"github.com/spf13/cobra"
)

func Execute(args []string) int {
	cmd := New(os.Stdout, os.Stderr)
	cmd.SetArgs(args)
	if err := fang.Execute(
		context.Background(),
		cmd,
		fang.WithoutVersion(),
		fang.WithErrorHandler(func(io.Writer, fang.Styles, error) {}),
	); err != nil {
		_ = render.WriteError(os.Stdout, os.Stderr, outputMode(cmd, args), err)
		return 1
	}
	return 0
}

func outputMode(cmd *cobra.Command, args []string) contract.OutputMode {
	if cmd != nil {
		if jsonOutput, err := cmd.Flags().GetBool("json"); err == nil && jsonOutput {
			return contract.OutputJSON
		}
	}
	for _, arg := range args {
		if arg == "--json" {
			return contract.OutputJSON
		}
	}
	return contract.OutputText
}
