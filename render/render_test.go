package render_test

import (
	"bytes"
	"testing"

	"github.com/vietmpl/vie/parse"
	"github.com/vietmpl/vie/render"
)

var tests = map[string]struct {
	source   string
	expected string
}{
	"text": {
		"text",
		"text",
	},
	"display-string-literal": {
		"{{ \"str\" }}",
		"str",
	},
	"display-pipe": {
		"{{ \"str\" | @upper }}",
		"STR",
	},
	"display-call": {
		"{{ @upper(\"str\") }}",
		"STR",
	},
	"not-false": {
		"{% if !false %}1{% end %}",
		"1",
	},
	"and": {
		"{% if true and true %}1{% end %}",
		"1",
	},
	"if-then-branch": {
		"{% if true %}1{% end %}",
		"1",
	},
	"if-else-branch": {
		"{% if false %}1{% else %}2{% end %}",
		"2",
	},
	"if-elseif-branch": {
		"{% if false %}1{% elseif true %}2{% end %}",
		"2",
	},
}

func TestSource(t *testing.T) {
	t.Parallel()
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			template, err := parse.Source([]byte(test.source))
			if err != nil {
				t.Fatal(err)
			}

			var buf bytes.Buffer
			render.MustRenderTemplate(&buf, template, nil)
			if test.expected != buf.String() {
				t.Fatalf("expected %q, got %q", test.expected, buf.String())
			}
		})
	}
}
