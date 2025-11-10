package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vietmpl/vie/parser"
	"github.com/vietmpl/vie/render"
	"github.com/vietmpl/vie/value"
)

func renderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "render <path>",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			params := make(map[string]value.Value)
			for _, a := range args[1:] {
				if strings.Contains(a, "=") {
					kv := strings.SplitN(a, "=", 2)
					params[kv[0]] = value.String(kv[1])
				} else {
					params[a] = value.Bool(true)
				}
			}

			sf := parser.Source(src)
			out, err := render.File(sf, params)
			if err != nil {
				return err
			}

			fmt.Fprint(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
	return cmd
}
