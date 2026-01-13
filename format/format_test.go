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
		"inline comment",
		"a{# comment #}b",
	},
	{
		"empty comment",
		"{##}",
	},
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
	name           string
	source         string
	expectedSource string
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
			expectFormat(t, test.source, test.source)
		})
	}

	for _, test := range transformTests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			expectFormat(t, test.source, test.expectedSource)
		})
	}
}

func expectFormat(t *testing.T, source, expectedSource string) {
	t.Helper()

	actual, err := format.Source([]byte(source))
	if err != nil {
		t.Error(err)
	}

	if expectedSource != string(actual) {
		t.Errorf("expected %q, got %q", expectedSource, actual)
	}
}
