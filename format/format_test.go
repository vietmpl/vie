package format_test

import (
	"testing"

	"github.com/vietmpl/vie/format"
	"github.com/vietmpl/vie/golden"
	"github.com/vietmpl/vie/parser"
)

func TestSource(t *testing.T) {
	t.Parallel()
	golden.RunGoldenTestdata(t, func(t *testing.T, input []byte) []byte {
		t.Parallel()
		sf := parser.ParseFile(input)
		actual := format.Source(sf)
		return actual
	})
}
