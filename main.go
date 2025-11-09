package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"github.com/vietmpl/vie/analisys"
	"github.com/vietmpl/vie/format"
	"github.com/vietmpl/vie/parser"
	"github.com/vietmpl/vie/render"
	"github.com/vietmpl/vie/value"
)

// Populated by goreleaser during build
var version = "unknown"

func main() {
	root := &cobra.Command{
		Use:     "vie",
		Version: version,
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}

	root.AddCommand(
		formatCmd(),
		contextCmd(),
		renderCmd(),
	)

	if err := fang.Execute(
		context.Background(),
		root,
		fang.WithVersion(root.Version),
		fang.WithErrorHandler(func(w io.Writer, _ fang.Styles, err error) {
			_, _ = fmt.Fprintln(w, "vie:", err)
		}),
		fang.WithColorSchemeFunc(fang.AnsiColorScheme),
	); err != nil {
		os.Exit(1)
	}
}

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
			res := format.File(f)
			fmt.Fprint(cmd.OutOrStdout(), string(res))
			return nil
		},
	}
	return cmd
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

			sf := parser.Source(src)
			tm, diagnostics := analisys.File(sf)
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
