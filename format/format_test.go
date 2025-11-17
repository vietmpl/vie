package format_test

import (
	"bytes"
	"testing"

	"github.com/skewb1k/golden"

	"github.com/vietmpl/vie/format"
	"github.com/vietmpl/vie/parser"
)

func TestSource(t *testing.T) {
	t.Parallel()
	golden.Run(t, func(t *testing.T, input []byte) []byte {
		t.Parallel()

		f, err := parser.ParseBytes(input, "")
		if err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		if err := format.FormatFile(&buf, f); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	})
}
