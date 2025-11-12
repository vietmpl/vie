package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vietmpl/vie/analysis"
	"github.com/vietmpl/vie/parser"
)

func printDiagnostics(path string, diagnostics []analysis.Diagnostic) {
	for _, d := range diagnostics {
		pos := d.Pos()
		fmt.Printf("%s:%d:%d: %s\n", path, pos.Line, pos.Character, d.String())
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

			f, err := parser.ParseBytes(src)
			if err != nil {
				return err
			}

			tm, diagnostics := analysis.CheckFile(f)
			if len(diagnostics) > 0 {
				printDiagnostics(path, diagnostics)
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
