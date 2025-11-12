package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

func Execute(ctx context.Context, use string, version string) error {
	root := &cobra.Command{
		Use:     use,
		Version: version,
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}

	root.AddCommand(
		formatCmd(),
		contextCmd(),
		renderCmd(),
	)

	root.SetVersionTemplate("{{.Version}}\n")

	root.InitDefaultVersionFlag()
	root.Flag("version").Usage = "Print version and exit"

	return fang.Execute(
		ctx,
		root,
		fang.WithVersion(root.Version),
		fang.WithErrorHandler(func(w io.Writer, _ fang.Styles, err error) {
			_, _ = fmt.Fprintln(w, err)
		}),
		fang.WithColorSchemeFunc(fang.AnsiColorScheme),
	)
}
