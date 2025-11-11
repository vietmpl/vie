package render_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/vietmpl/vie/parser"
	"github.com/vietmpl/vie/pkg/golden"
	"github.com/vietmpl/vie/render"
	"github.com/vietmpl/vie/value"
)

func TestSource(t *testing.T) {
	t.Parallel()

	type testCase struct {
		context map[string]value.Value
	}

	cases := map[string]testCase{
		"var.vie": {
			context: map[string]value.Value{
				"name":        value.String("test"),
				"emptystring": value.String(""),
				"false_flag":  value.Bool(false),
				"true_flag":   value.Bool(true),
				"switch":      value.String("123"),
			},
		},
		"call-var.vie": {
			context: map[string]value.Value{
				"name": value.String("test"),
			},
		},
		"pipe-var.vie": {
			context: map[string]value.Value{
				"name": value.String("test"),
			},
		},
	}

	golden.Run(t, func(t *testing.T, input []byte) []byte {
		t.Parallel()

		f := parser.Source(input)

		name := strings.TrimPrefix(t.Name(), "TestSource/")
		context := cases[name].context

		var b bytes.Buffer
		render.MustRenderFile(&b, f, context)
		return b.Bytes()
	})
}
