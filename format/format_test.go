package format_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vietmpl/vie/format"
	"github.com/vietmpl/vie/parser"
)

func TestGoldenFiles(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), ".golden") {
			continue
		}

		name := e.Name()
		t.Run(name, func(t *testing.T) {
			inputPath := filepath.Join("testdata", name)
			goldenPath := inputPath + ".golden"

			input, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatal(err)
			}

			sf := parser.ParseFile(input)

			got := format.Source(sf)

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("missing golden file: %v", err)
			}

			if !bytes.Equal(got, want) {
				t.Errorf("mismatch for %s\n--- got\n%s\n--- want\n%s", name, got, want)
			}
		})
	}
}
