package main

import (
	"context"
	"os"

	"github.com/vietmpl/vie/internal/cli"
)

var use = "vie"
var version = "v0.0.1"

func main() {
	if err := cli.Execute(context.Background(), use, version); err != nil {
		os.Exit(1)
	}
}
