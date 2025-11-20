package cli

import (
	"github.com/spf13/cobra"
	"github.com/vietmpl/vie/internal/lsp"
)

func newCmdLSP() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "lsp",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			server := lsp.NewServer()
			server.RunStdio()
			return nil
		},
	}
	return cmd
}
