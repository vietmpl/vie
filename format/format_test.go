package format_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vietmpl/vie/format"
	"github.com/vietmpl/vie/parser"
)

func TestSource(t *testing.T) {
	tests, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		if strings.HasSuffix(test.Name(), ".golden") {
			continue
		}

		name := test.Name()
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

			assert.Equal(t, string(got), string(want))
		})
	}
}
