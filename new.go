package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vietmpl/vie/template"
)

func newCmdNew() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "new TEMPLATE DEST [VAR=VALUE...] [VAR...]",
		Short:   "Render a template in the target directory",
		Example: "vie new component src/ name=Button with-test",
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			tmplName := args[0]
			dest := args[1]

			tmplPath := filepath.Join(".vie", tmplName)
			tmpl, err := template.FromDir(tmplPath)
			if err != nil {
				return err
			}

			dataArgs := args[2:]
			data, err := parseData(dataArgs)
			if err != nil {
				return err
			}

			files, err := tmpl.Render(data)
			if err != nil {
				return err
			}

			exit := false
			for name := range files {
				path := filepath.Join(dest, name)
				if _, err := os.Stat(path); err == nil {
					exit = true
					fmt.Printf("failed to create %q: file already exists\n", path)
				}
			}
			if exit {
				os.Exit(1)
			}

			for relPath, content := range files {
				path := filepath.Join(dest, relPath)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(path, content, 0o644); err != nil {
					return err
				}
				fmt.Println(path)
			}
			return nil
		},
	}
	return cmd
}
