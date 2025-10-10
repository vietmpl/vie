package format_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vietmpl/vie/format"
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
				var b strings.Builder
				b.Grow(len(input))
				if err := format.Source(&b, []byte(input)); err != nil {
					t.Fatal(err)
				}
				got := b.String()
				if got != expected {
					// TODO: show diff
					t.Errorf("%s\n--- expected ---\n%s\n--- got ---\n%s",
						n.Name(), expected, got)
				}
			})
		}
	}
}

// loadTestFile parses testdata file with --- delimiters
func loadTestFile(t *testing.T, path string) (string, []string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(string(data), "---")
	if len(parts) < 1 {
		t.Fatalf("%s: test must have expected output and at least one input", path)
	}
	var inputs []string
	expected := strings.TrimSpace(parts[0])
	for _, p := range parts[1:] {
		inputs = append(inputs, strings.TrimSpace(p))
	}
	return expected, inputs
}
