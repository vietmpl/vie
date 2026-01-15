package render_test

import (
	"testing"

	"github.com/vietmpl/vie/parse"
	"github.com/vietmpl/vie/render"
	"github.com/vietmpl/vie/value"
)

var noContextTests = map[string]struct {
	source         string
	expectedSource string
}{
	"text": {
		"text",
		"text",
	},
	"comment": {
		"{# #}",
		"",
	},
	"display string literal": {
		"{{ \"str\" }}",
		"str",
	},
	"display undefined variable": {
		"{{ undefined }}",
		"",
	},
	"display call": {
		"{{ @upper(\"str\") }}",
		"STR",
	},
	"display pipe": {
		"{{ \"str\" | @upper }}",
		"STR",
	},
	"not false": {
		"{% if !false %}1{% end %}",
		"1",
	},
	"and": {
		"{% if true and true %}1{% end %}",
		"1",
	},
	"or": {
		"{% if false or true %}1{% end %}",
		"1",
	},
	"precedence": {
		"{% if true or true and false %}1{% end %}",
		"1",
	},
	"precedence with parens": {
		"{% if (true or true) and false %}1{% end %}",
		"",
	},
	"if else branch": {
		"{% if false %}1{% else %}2{% end %}",
		"2",
	},
	"if elseif branch": {
		"{% if false %}1{% elseif true %}2{% end %}",
		"2",
	},
	"if undefined": {
		"{% if undefined %}1{% end %}",
		"",
	},
	"multiline if": {
		"{% if true %}\n1\n{% end %}\n",
		"\n1\n\n",
	},
	"false equal false": {
		"{% if false == false %}1{% end %}",
		"1",
	},
	"false not equal true": {
		"{% if false != true %}1{% end %}",
		"1",
	},
	"empty string equal empty string": {
		"{% if \"\" == \"\" %}1{% end %}",
		"1",
	},
	"empty string not equal empty string": {
		"{% if \"a\" != \"b\" %}1{% end %}",
		"1",
	},
}

var contextTests = map[string]struct {
	source         string
	expectedSource string
	context        map[string]value.Value
}{
	"display var": {
		"{{ a }}",
		"foo",
		map[string]value.Value{
			"a": value.String("foo"),
		},
	},
}

var errorTests = map[string]struct {
	source  string
	context map[string]value.Value
}{
	"display bool": {
		"{{ true }}",
		nil,
	},
	"display not bool": {
		"{{ !true }}",
		nil,
	},
	"concatenate bools": {
		"{{ false ~ true }}",
		nil,
	},
	"call to undefined function": {
		"{{ @undefined()}}",
		nil,
	},
	"call with missing argument": {
		"{{ @upper() }}",
		nil,
	},
	"call with extra argument": {
		"{{ @upper(\"\", \"\") }}",
		nil,
	},
	"call with wrong type": {
		"{{ @upper(true) }}",
		nil,
	},
	"pipe with wrong type": {
		"{{ true | @upper }}",
		nil,
	},
	"display or": {
		"{{ false or true }}",
		nil,
	},
	"display and": {
		"{{ false and true }}",
		nil,
	},
	"display equal": {
		"{{ false == true }}",
		nil,
	},
	"display not equal": {
		"{{ false != true }}",
		nil,
	},
	"compare bool and string": {
		"{% if false == \"\" %}{% end %}",
		nil,
	},
	"compare string and bool": {
		"{% if \"\" == false %}{% end %}",
		nil,
	},
}

func TestTemplate(t *testing.T) {
	t.Parallel()

	for name, test := range noContextTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			expectRender(t, test.source, nil, test.expectedSource)
		})
	}

	for name, test := range contextTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			expectRender(t, test.source, test.context, test.expectedSource)
		})
	}

	for name, test := range errorTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			template, err := parse.Source([]byte(test.source))
			if err != nil {
				t.Error(err)
			}

			if _, err := render.Template(template, test.context); err == nil {
				t.Errorf("succeeded unexpectedly")
			}
		})
	}
}

func expectRender(
	t *testing.T,
	source string,
	context map[string]value.Value,
	expectedSource string,
) {
	t.Helper()

	template, err := parse.Source([]byte(source))
	if err != nil {
		t.Fatal(err)
	}

	actual, err := render.Template(template, context)
	if err != nil {
		t.Error(err)
	}

	if expectedSource != string(actual) {
		t.Errorf("expected %q, got %q", expectedSource, actual)
	}
}
