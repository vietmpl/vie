package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vietmpl/vie/analysis"
	"github.com/vietmpl/vie/parser"
	"github.com/vietmpl/vie/render"
	"github.com/vietmpl/vie/value"
)

func parseContext(args []string) (map[string]value.Value, error) {
	context := make(map[string]value.Value)
	for _, a := range args {
		if strings.Contains(a, "=") {
			kv := strings.SplitN(a, "=", 2)
			context[kv[0]] = value.String(kv[1])
		} else {
			context[a] = value.Bool(true)
		}
	}
	return context, nil
}

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

			f := parser.Source(src)

			c, err := parseContext(args[1:])
			if err != nil {
				return err
			}

			tm, diagnostics := analysis.File(f)
			if len(diagnostics) > 0 {
				printDiagnostics(cmd.OutOrStdout(), path, diagnostics)
				return nil
			}

			exit := false

			for varname, typ := range tm {
				val, ok := c[varname]
				if ok {
					if val.Type() != typ {
						fmt.Fprintf(cmd.OutOrStdout(), "%s: expected %s, got %s\n", varname, typ, val.Type())
						exit = true
					}
				}
			}

			if exit {
				return nil
			}

			// Variables in context but not in template
			for varname := range c {
				if _, ok := tm[varname]; !ok {
					fmt.Fprintf(cmd.OutOrStdout(), "warning: %s provided but not used\n", varname)
				}
			}

			out, err := render.File(f, c)
			if err != nil {
				return err
			}

			fmt.Fprint(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
	return cmd
}
