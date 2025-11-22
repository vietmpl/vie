package format_test

import (
	"testing"

	"github.com/skewb1k/golden"

	"github.com/vietmpl/vie/format"
)

func TestSource(t *testing.T) {
	t.Parallel()
	golden.Run(t, func(t *testing.T, input []byte) []byte {
		t.Parallel()
		result, err := format.Source(input)
		if err != nil {
			t.Fatal(err)
		}
		return result
	})
}
