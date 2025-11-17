package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vietmpl/vie/analysis"
	"github.com/vietmpl/vie/parser"
)

func printDiagnostics(diagnostics []analysis.Diagnostic) {
	for _, d := range diagnostics {
		pos := d.Pos()
		fmt.Printf("%s:%d:%d: %s\n", pos.Path, pos.Line, pos.Character, d.String())
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

			f, err := parser.ParseBytes(src, path)
			if err != nil {
				return err
			}

			analyzer := analysis.NewAnalyzer()
			analyzer.File(f)
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
