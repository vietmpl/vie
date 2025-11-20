package render_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/skewb1k/golden"

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

		f, err := parser.ParseBytes(input)
		if err != nil {
			t.Fatal(err)
		}

		name := strings.TrimPrefix(t.Name(), "TestSource/")
		context := cases[name].context

		var buf bytes.Buffer
		render.MustRenderFile(&buf, f, context)
		return buf.Bytes()
	})
}
