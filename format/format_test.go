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
		"inline comment",
		"a{# comment #}b",
	},
	{
		"empty comment",
		"{# #}",
	},
	{
		"multiline if block",
		"a\n{% if true %}\nb\n{% end %}\nc",
	},
}

var transformTests = [...]struct {
	name            string
	source          string
	expected_source string
}{
	{
		"inline if expands to multiline",
		"a{% if true %}b{% end %}c",
		"a\n{% if true %}\nb\n{% end %}\nc",
	},
	{
		"trailing whitespace is removed",
		"a\n{% if true %}  \nb\n{% end %} \nc",
		"a\n{% if true %}\nb\n{% end %}\nc",
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
