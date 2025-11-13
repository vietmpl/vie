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
	var stdio bool

	cmd := &cobra.Command{
		Use:  "format <path>",
		Args: cobra.MaximumNArgs(1), // 0 if --stdio, 1 otherwise
		RunE: func(cmd *cobra.Command, args []string) error {
			if stdio {
				src, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				formatted, err := formatBytes(src)
				if err != nil {
					return err
				}
				if check {
					if !bytes.Equal(src, formatted) {
						os.Exit(1)
					}
					return nil
				} else {
					_, err = os.Stdout.Write(formatted)
					return err
				}
			}

			if len(args) != 1 {
				return fmt.Errorf("expected at least one file or directory argument")
			}
			path := args[0]

			info, err := os.Stat(path)
			if err != nil {
				return err
			}

			if !info.IsDir() {
				changed, err := formatFile(path, check)
				if err != nil {
					return err
				}
				if check && changed {
					os.Exit(1)
				}
			}

			changed := false
			// TODO(skewb1k): consider parallelizing formatting for multiple files.
			// Might not be worth the added complexity; using a buffer pool
			// could be more appealing in this case.
			if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || filepath.Ext(d.Name()) != ".vie" {
					return nil
				}
				c, err := formatFile(path, check)
				if err != nil {
					return err
				}
				changed = changed || c
				return nil
			}); err != nil {
				return err
			}
			if changed {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&check, "check", "c", false, "List non-conforming files and exit with an error if the list is non-empty")
	cmd.Flags().BoolVar(&stdio, "stdio", false, "Read input from stdin and write formatted output to stdout")

	return cmd
}

func formatFile(path string, check bool) (changed bool, err error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return
	}
	defer f.Close()
	src, err := io.ReadAll(f)
	if err != nil {
		return
	}
	formatted, err := formatBytes(src)
	if err != nil {
		return
	}

	// Ignore if content has not changed
	if bytes.Equal(src, formatted) {
		return
	}
	changed = true
	// Print path of changed file
	fmt.Println(path)

	if !check {
		// Truncate and seek to start before rewriting the file
		if err = f.Truncate(0); err != nil {
			return
		}
		if _, err = f.Seek(0, 0); err != nil {
			return
		}
		if _, err = f.Write(formatted); err != nil {
			return
		}
	}
	return
}

// TODO(skewb1k): consider moving this under the vie/format package.
func formatBytes(src []byte) ([]byte, error) {
	parsed, err := parser.ParseBytes(src)
	if err != nil {
		return nil, err
	}
	// Bufferize formatted output to catch errors before touching the file and
	// compare with the original source.
	var buf bytes.Buffer
	// Preallocate buffer to reduce memory allocations.
	// Heuristic: half of the original file size.
	buf.Grow(len(src) / 2)
	if err := format.FormatFile(&buf, parsed); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
