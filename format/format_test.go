package format_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vietmpl/vie/format"
	"github.com/vietmpl/vie/parser"
)

func TestFormatting(t *testing.T) {
	base := "testdata"
	root, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range root {
		expected, inputs := loadTestFile(t, filepath.Join(base, n.Name()))
		for i, input := range inputs {
			t.Run(fmt.Sprintf("case-%d", i+1), func(t *testing.T) {

				sourceFile, err := parser.ParseFile(input)
				if err != nil {
					t.Fatal(err)
				}

				got := format.Source(sourceFile)
				if !bytes.Equal(got, expected) {
					// TODO: show diff
					t.Errorf("%s\n--- expected ---\n%s\n--- got ---\n%s",
						n.Name(), expected, got)
				}
			})
		}
	}
}

// loadTestFile parses testdata file with --- delimiters
func loadTestFile(t *testing.T, path string) ([]byte, [][]byte) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(string(data), "---")
	if len(parts) < 1 {
		t.Fatalf("%s: test must have expected output and at least one input", path)
	}
	var inputs [][]byte
	for _, p := range parts[1:] {
		inputs = append(inputs, []byte(strings.TrimSpace(p)))
	}
	expected := []byte(strings.TrimSpace(parts[0]))
	return expected, inputs
}
