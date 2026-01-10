package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vietmpl/vie/analysis"
	"github.com/vietmpl/vie/parse"
)

func newCmdContext() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "context PATH",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			f, err := parse.Source(src)
			if err != nil {
				return err
			}

			analyzer := analysis.NewAnalyzer()
			analyzer.Template(f, path)
			tm, diagnostics := analyzer.Results()
			if diagnostics != nil {
				printDiagnostics(diagnostics)
				return nil
			}
			// TODO(skewb1k): improve output format.
			for varname, typ := range tm {
				fmt.Printf("%s: %s\n", varname, typ.String())
			}
			return nil
		},
	}
	return cmd
}

func printDiagnostics(diagnostics []analysis.Diagnostic) {
	for _, d := range diagnostics {
		pos := d.Pos()
		fmt.Printf("%s:%d:%d: %s\n", d.Path(), pos.Line, pos.Column, d.String())
	}
}
