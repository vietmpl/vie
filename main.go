package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var version = "v0.0.1"

func main() {
	log.SetFlags(0)
	log.SetPrefix("error: ")

	root := cobra.Command{
		Use:          "vie",
		Version:      version,
		SilenceUsage: true,
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}
	root.SetErrPrefix("error:")

	root.SetVersionTemplate("{{.Version}}\n")
	root.InitDefaultVersionFlag()
	root.Flag("version").Usage = "Print version and exit"

	root.AddCommand(
		newCmdFormat(),
		newCmdContext(),
		newCmdRender(),
		newCmdNew(),
		newCmdList(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
