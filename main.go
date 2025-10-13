package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"github.com/vietmpl/vie/analisys"
	"github.com/vietmpl/vie/format"
	"github.com/vietmpl/vie/parser"
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
		contextCmd(),
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

			sourceFile, err := parser.ParseFile(src)
			res := format.Source(sourceFile)
			fmt.Fprint(cmd.OutOrStdout(), string(res))
			return nil
		},
	}
	return cmd
}

func contextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "context <path>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			sourceFile, err := parser.ParseFile(src)
			tm, diagnostics := analisys.Source(sourceFile)
			if len(diagnostics) != 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "%v\n", diagnostics)
			}

			for varname, typ := range tm {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", varname, typ.String())
			}
			return nil
		},
	}
	return cmd
}

func errorHandler(w io.Writer, _ fang.Styles, err error) {
	_, _ = fmt.Fprintln(w, "vie:", err)
}
