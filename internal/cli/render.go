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

// checkForUnusedVars checks for variables in the context but not in template.
func checkForUnusedVars(typemap map[string]value.Type, context map[string]value.Value) {
	for varname := range context {
		if _, ok := typemap[varname]; !ok {
			fmt.Printf("warning: %s provided but not used\n", varname)
		}
	}
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

			f, err := parser.ParseBytes(src, path)
			if err != nil {
				return err
			}

			context, err := parseContext(args[1:])
			if err != nil {
				return err
			}

			analyzer := analysis.NewAnalyzer()
			analyzer.File(f)
			typemap, diagnostics := analyzer.Results()
			if len(diagnostics) > 0 {
				printDiagnostics(diagnostics)
				return nil
			}

			errs := analysis.MergeTypes(typemap, context)
			if len(errs) > 0 {
				for _, err := range errs {
					fmt.Println(err)
				}
				return nil
			}

			checkForUnusedVars(typemap, context)

			render.MustRenderFile(os.Stdout, f, context)
			return nil
		},
	}
	return cmd
}
