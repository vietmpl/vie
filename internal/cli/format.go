package cli

import (
	"bytes"
	"fmt"
	"io"
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

			if !info.IsDir() {
				return formatFile(arg, check)
			}

			// TODO(skewb1k): consider parallelizing formatting for multiple files.
			// Might not be worth the added complexity; using a buffer pool
			// could be more appealing in this case.
			return filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || filepath.Ext(d.Name()) != ".vie" {
					return nil
				}
				return formatFile(path, check)
			})
		},
	}

	cmd.Flags().BoolVarP(&check, "check", "c", false, "Don't write format result back to the source file")

	return cmd
}

func formatFile(path string, check bool) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	src, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	parsed, err := parser.ParseBytes(src)
	if err != nil {
		return err
	}

	// Bufferize formatted output to catch errors before touching the file and
	// compare with the original source.
	var buf bytes.Buffer
	// Preallocate buffer to reduce memory allocations.
	// Heuristic: half of the original file size.
	buf.Grow(len(src) / 2)

	if err := format.FormatFile(&buf, parsed); err != nil {
		return err
	}
	formatted := buf.Bytes()

	// Check if content has changed
	if !bytes.Equal(formatted, src) {
		// Print path of changed file
		fmt.Println(path)
		if !check {
			// Truncate and seek to start before rewriting the file
			if err := f.Truncate(0); err != nil {
				return err
			}
			if _, err := f.Seek(0, 0); err != nil {
				return err
			}
			if _, err := f.Write(formatted); err != nil {
				return err
			}
		}
	}
	return nil
}
