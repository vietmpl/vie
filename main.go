package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

// Populated by goreleaser during build
var version = "unknown"

func main() {
	root := &cobra.Command{
		Use:     "vie",
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

	if err := fang.Execute(
		context.Background(),
		root,
		fang.WithVersion(root.Version),
		fang.WithErrorHandler(func(w io.Writer, _ fang.Styles, err error) {
			_, _ = fmt.Fprintln(w, "vie:", err)
		}),
		fang.WithColorSchemeFunc(fang.AnsiColorScheme),
	); err != nil {
		os.Exit(1)
	}
}
