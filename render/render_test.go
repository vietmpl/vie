package render_test

import (
	"testing"

	"github.com/vietmpl/vie/golden"
	"github.com/vietmpl/vie/parser"
	"github.com/vietmpl/vie/render"
)

func TestSource(t *testing.T) {
	golden.Run(t, func(t *testing.T, input []byte) []byte {
		sf := parser.ParseFile(input)
		actual, err := render.Source(sf, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return actual
	})
}
