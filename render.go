package main

import (
	"bytes"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vietmpl/vie/parse"
	"github.com/vietmpl/vie/render"
	"github.com/vietmpl/vie/value"
)

func newCmdRender() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "render PATH [VAR=VALUE...] [VAR...]",
		Args: cobra.MinimumNArgs(1),
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

			context, err := parseContext(args[1:])
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			if err := render.Template(&buf, f, context); err != nil {
				return err
			}
			_, err = os.Stdout.Write(buf.Bytes())
			return err
		},
	}
	return cmd
}

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
