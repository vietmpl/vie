package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"vie": main,
	})
}

func Test(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testscript.Run(t, testscript.Params{
				Dir: filepath.Join("testdata", name),
			})
		})
	}
}
