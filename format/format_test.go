package format_test

import (
	"bytes"
	"testing"

	"github.com/vietmpl/vie/format"
	"github.com/vietmpl/vie/parser"
	"github.com/vietmpl/vie/pkg/golden"
)

func TestSource(t *testing.T) {
	t.Parallel()
	golden.Run(t, func(t *testing.T, input []byte) []byte {
		t.Parallel()
		f := parser.Source(input)
		var b bytes.Buffer
		if err := format.FormatFile(&b, f); err != nil {
			t.Fatal(err)
		}
		return b.Bytes()
	})
}
