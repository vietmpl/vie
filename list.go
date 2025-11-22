package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List available templates",
		Args:    cobra.NoArgs,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			entries, err := os.ReadDir(".vie")
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return nil
				}
				return err
			}
			for _, e := range entries {
				if e.IsDir() {
					fmt.Println(e.Name())
				}
			}
			return nil
		},
	}
	return cmd
}
