package format_test

import (
	"testing"

	"github.com/vietmpl/vie/format"
	"github.com/vietmpl/vie/parser"
	"github.com/vietmpl/vie/pkg/golden"
)

func TestSource(t *testing.T) {
	t.Parallel()
	golden.Run(t, func(t *testing.T, input []byte) []byte {
		t.Parallel()
		sf := parser.Source(input)
		actual := format.File(sf)
		return actual
	})
}
