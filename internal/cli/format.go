package cli

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vietmpl/vie/format"
	"github.com/vietmpl/vie/parser"
)

func formatCmd() *cobra.Command {
	var check bool

	cmd := &cobra.Command{
		Use:  "format <path>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg := args[0]

			info, err := os.Stat(arg)
			if err != nil {
				return err
			}

			formatFile := func(path string) error {
				f, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				parsed, err := parser.ParseBytes(f)
				if err != nil {
					return err
				}

				var buf bytes.Buffer
				if err := format.FormatFile(&buf, parsed); err != nil {
					return err
				}
				formatted := buf.Bytes()

				// Check if content has changed
				if !bytes.Equal(formatted, f) {
					fmt.Println(path)
					if !check {
						if err := os.WriteFile(path, formatted, info.Mode()); err != nil {
							return err
						}
					}
				}
				return nil
			}

			if !info.IsDir() {
				return formatFile(arg)
			}

			return filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || filepath.Ext(d.Name()) != ".vie" {
					return nil
				}
				return formatFile(path)
			})
		},
	}

	cmd.Flags().BoolVarP(&check, "check", "c", false, "Don't write format result back to the source file")

	return cmd
}
