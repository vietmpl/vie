package format_test

import (
	"testing"

	"github.com/vietmpl/vie/format"
)

var stableTests = [...]struct {
	name   string
	source string
}{
	{
		"empty",
		"",
	},
	{
		"single newline",
		"\n",
	},
	{
		"display block",
		"a{{ id }}b",
	},
	{
		"pipe",
		"{{ foo | bar }}",
	},
	{
		"binary operators",
		"{{ a ~ b or c == d and e != g }}",
	},
	{
		"unary operators",
		"{{ !a }}",
	},
	{
		"parens",
		"{{ ((a)) }}",
	},
	{
		"conditional",
		"{{ true ? a : b }}",
	},
	{
		"inline comment",
		"a{# comment #}b",
	},
	{
		"empty comment",
		"{##}",
	},
	// {
	// 	"multiline comment",
	// 	"{# 1\n2\n3#}",
	// },
	{
		"single space comment",
		"{# #}",
	},
	{
		"multiline if block",
		"a\n{% if true %}\nb\n{% end %}\nc",
	},
	{
		"inline if",
		"{% if a %}b{% elseif b %}{% else %}{% end %}",
	},
	{
		"multiple elseif's",
		"{% if a %}\n{% elseif b %}\n{% elseif c %}\n{% else %}\n{% end %}",
	},
	{
		"trailing whitespace",
		"\n{% if true %} \n\n{% end %}\n",
	},
}

var transformTests = [...]struct {
	name            string
	source          string
	expected_source string
}{
	{
		"adds spaces around display",
		"{{name}}",
		"{{ name }}",
	},
	{
		"fix spaces around statements",
		"{% 	if name  %}{%  end   %}",
		"{% if name %}{% end %}",
	},
}

func TestSource(t *testing.T) {
	t.Parallel()

	for _, test := range stableTests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testFormat(t, test.source, test.source)
		})
	}

	for _, test := range transformTests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testFormat(t, test.source, test.expected_source)
		})
	}
}

func testFormat(t *testing.T, source, expected_source string) {
	t.Helper()

	actual, err := format.Source([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	if expected_source != string(actual) {
		t.Fatalf("expected %q, got %q", expected_source, actual)
	}
}
