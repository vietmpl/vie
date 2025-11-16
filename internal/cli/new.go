package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	cli "github.com/vietmpl/vie/internal/cli/lib"
	"github.com/vietmpl/vie/internal/template"
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

func newCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "new <template-name | path> <dest>",
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dist, err := os.Getwd()
			if err != nil {
				return err
			}

			currentPath, err := filepath.Abs(dist)
			if err != nil {
				return err
			}

			root := filepath.VolumeName(currentPath)
			if root == "" {
				root = "/"
			} else {
				root = root + string(filepath.Separator)
			}

			fsys := os.DirFS(root)

			relPath, err := filepath.Rel(root, currentPath)
			if err != nil {
				return err
			}
			if relPath == "." {
				relPath = ""
			}

			result := cli.SearchTemplate(fsys, relPath, args[0])
			path := args[0]
			if result != "" {
				path = filepath.Join(root, result)
			}

			dest := args[1]

			tmpl, err := template.FromDir(path)
			if err != nil {
				return err
			}

			context, err := parseContext(args[2:])
			if err != nil {
				return err
			}

			files, err := tmpl.Render(context)
			if err != nil {
				return err
			}

			for name := range files {
				path := filepath.Join(dest, name)
				if _, err := os.Stat(path); err == nil {
					return fmt.Errorf("failed to create %q: file already exists", path)
				}
			}

			for relPath, content := range files {
				path := filepath.Join(dest, relPath)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(path, content, 0o644); err != nil {
					return err
				}
				fmt.Println(relPath)
			}
			return nil
		},
	}
	return cmd
}
