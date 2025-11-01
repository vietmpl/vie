package format_test

import (
	"testing"

	"github.com/vietmpl/vie/format"
	"github.com/vietmpl/vie/golden"
	"github.com/vietmpl/vie/parser"
)

func TestSource(t *testing.T) {
	golden.Run(t, func(t *testing.T, input []byte) []byte {
		sf := parser.ParseFile(input)
		res := format.Source(sf)
		return res
	})
}
