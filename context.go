package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/vietmpl/vie/analysis"
	"github.com/vietmpl/vie/parser"
)

func printDiagnostics(w io.Writer, path string, diagnostics []analysis.Diagnostic) {
	for _, d := range diagnostics {
		pos := d.Pos()
		fmt.Fprintf(w, "%s:%d:%d: %s\n", path, pos.Line, pos.Character, d.String())
	}
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

			f := parser.Source(src)

			tm, diagnostics := analysis.File(f)
			if len(diagnostics) > 0 {
				printDiagnostics(cmd.OutOrStdout(), path, diagnostics)
				return nil
			}
			for varname, typ := range tm {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", varname, typ.String())
			}
			return nil
		},
	}
	return cmd
}
