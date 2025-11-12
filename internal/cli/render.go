package cli

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

			f, err := parser.ParseBytes(src)
			if err != nil {
				return err
			}

			c, err := parseContext(args[1:])
			if err != nil {
				return err
			}

			tm, diagnostics := analysis.CheckFile(f)
			if len(diagnostics) > 0 {
				printDiagnostics(path, diagnostics)
				return nil
			}

			exit := false
			for varname, typ := range tm {
				val, ok := c[varname]
				if ok {
					if val.Type() != typ {
						fmt.Printf("%s: expected %s, got %s\n", varname, typ, val.Type())
						exit = true
					}
				} else if !exit {
					// Assign a default value for missing variables.
					// TODO(skewb1k): maybe the parser should handle undefined variables.
					switch typ {
					case value.TypeBool:
						c[varname] = value.Bool(false)
					case value.TypeString:
						c[varname] = value.String("")
					default:
						panic(fmt.Sprintf("unexpected Type value: %d", typ))
					}
				}
			}
			if exit {
				return nil
			}

			// Variables in context but not in template
			for varname := range c {
				if _, ok := tm[varname]; !ok {
					fmt.Printf("warning: %s provided but not used\n", varname)
				}
			}

			render.MustRenderFile(os.Stdout, f, c)
			return nil
		},
	}
	return cmd
}
