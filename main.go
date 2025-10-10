package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"github.com/vietmpl/vie/format"
)

// Populated by goreleaser during build
var version = "unknown"

func main() {
	root := &cobra.Command{
		Use:     "vie",
		Version: version,
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}

	root.AddCommand(
		formatCmd(),
	)

	if err := fang.Execute(
		context.Background(),
		root,
		fang.WithVersion(root.Version),
		fang.WithErrorHandler(errorHandler),
		fang.WithColorSchemeFunc(fang.AnsiColorScheme),
	); err != nil {
		os.Exit(1)
	}
}

func formatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "format <path>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var buf bytes.Buffer

			if err := format.Source(&buf, src); err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), buf.String())
			return nil
		},
	}
	return cmd
}

func errorHandler(w io.Writer, _ fang.Styles, err error) {
	_, _ = fmt.Fprintln(w, "vie:", err)
}
