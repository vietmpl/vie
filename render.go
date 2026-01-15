package main

import (
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

			data, err := parseData(args[1:])
			if err != nil {
				return err
			}

			out, err := render.Template(f, data)
			if err != nil {
				return err
			}
			_, err = os.Stdout.Write(out)
			return err
		},
	}
	return cmd
}

func parseData(args []string) (map[string]value.Value, error) {
	data := make(map[string]value.Value)
	for _, a := range args {
		if strings.Contains(a, "=") {
			kv := strings.SplitN(a, "=", 2)
			data[kv[0]] = value.String(kv[1])
		} else {
			data[a] = value.Bool(true)
		}
	}
	return data, nil
}
