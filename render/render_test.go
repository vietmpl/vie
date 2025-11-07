package render_test

import (
	"path/filepath"
	"testing"

	"github.com/vietmpl/vie/golden"
	"github.com/vietmpl/vie/parser"
	"github.com/vietmpl/vie/render"
	"github.com/vietmpl/vie/value"
)

func TestSource(t *testing.T) {
	t.Parallel()

	type testCase struct {
		context map[string]value.Value
	}

	cases := map[string]testCase{
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
		"TestSource/function/call-var.vie": {
			context: map[string]value.Value{
				"name": value.String("test"),
			},
		},
	}

	golden.Run(t, func(t *testing.T, input []byte) []byte {
		t.Parallel()
		sf := parser.ParseFile(input)

		name := filepath.ToSlash(t.Name())
		context := cases[name].context

		actual, err := render.Source(sf, context)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return actual
	})
}
