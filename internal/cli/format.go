package cli

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vietmpl/vie/format"
	"github.com/vietmpl/vie/parser"
)

// TODO(nickshiro): support dirs
func formatCmd() *cobra.Command {
	var check bool

	cmd := &cobra.Command{
		Use:  "format <path>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			src, err := os.ReadFile(path)
			info, err := os.Stat(path)
			if err != nil {
				return err
			}

			f, err := parser.ParseBytes(src)
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			if err := format.FormatFile(&buf, f); err != nil {
				return err
			}

			if check {
				fmt.Printf("%s", buf.Bytes())
			} else {
				if err := os.WriteFile(path, buf.Bytes(), info.Mode()); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&check, "check", "c", false, "Don't write format result back to the source file")

	return cmd
}
