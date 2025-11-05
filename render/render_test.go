package render_test

import (
	"testing"

	"github.com/vietmpl/vie/golden"
	"github.com/vietmpl/vie/parser"
	"github.com/vietmpl/vie/render"
	"github.com/vietmpl/vie/value"
)

func TestSource(t *testing.T) {
	type tc struct {
		context map[string]value.Value
	}

	cases := map[string]tc{
		// TODO(skewb1k): avoid specifing full path.
		"TestSource/var.vie": {
			context: map[string]value.Value{
				"name":        value.String("test"),
				"emptystring": value.String(""),
				"false_flag":  value.Bool(false),
				"true_flag":   value.Bool(true),
				"switch":      value.String("123"),
			},
		},
	}

	golden.Run(t, func(t *testing.T, input []byte) []byte {
		sf := parser.ParseFile(input)

		context := cases[t.Name()].context

		actual, err := render.Source(sf, context)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return actual
	})
}
