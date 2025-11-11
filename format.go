package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vietmpl/vie/format"
	"github.com/vietmpl/vie/parser"
)

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

			f := parser.Source(src)
			if err := format.FormatFile(os.Stdout, f); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}
