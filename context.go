package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vietmpl/vie/analysis"
	"github.com/vietmpl/vie/parser"
)

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

			sf := parser.Source(src)
			tm, diagnostics := analysis.File(sf)
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
