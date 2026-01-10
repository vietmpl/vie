package parse_test

import (
	"strconv"
	"testing"

	"github.com/vietmpl/vie/parse"
)

var errorTests = [...]struct {
	name   string
	source string
}{
	{
		"empty-tag",
		"{%%}",
	},
	{
		"empty-display",
		"{{ }}",
	},
	{
		"empty-if",
		"{% if %}{% end %}",
	},
	{
		"missing-pipe-function",
		"{{ true | }}",
	},
	{
		"missing-pipe-arg",
		"{{ | @upper }}",
	},
	{
		"missing-if-end",
		"{% if true %}",
	},
	{
		"trailing-end",
		"{% end %}",
	},
	{
		"unclosed-comment",
		"{#",
	},
	{
		"unclosed-string-literal",
		"{{ \" }}",
	},
	{
		"unclosed-conditional-expr",
		"{{ true ? \"\" }}",
	},
	{
		"unclosed-if-tag",
		"{% if true",
	},
	{
		"extra-display-expr",
		"{{ \"\" true }}",
	},
	{
		"extra-else-tag",
		"{% if true %}{% else %}{% else %}{% end %}",
	},
	{
		"multiline-display", "{{ \n }}",
	},
}

func TestSource(t *testing.T) {
	t.Parallel()
	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := parse.Source([]byte(test.source))
			if err == nil {
				t.Fatalf("expected error, got no error")
			}
		})
	}
}

// TestSourceFuzz checks that [parse.Source] never panics on incomplete or
// slightly modified input. It concatenates all [errorTests] sources, then
// tests truncating from the start, truncating from the end, and removing each
// character.
func TestSourceFuzz(t *testing.T) {
	t.Parallel()

	var combined []byte
	for _, test := range errorTests {
		combined = append(combined, test.source...)
	}

	// Truncate from start
	for i := range combined {
		t.Run("start_"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			fragment := combined[i:]
			_, _ = parse.Source(fragment)
		})
	}

	// Truncate from end
	for i := len(combined); i >= 0; i-- {
		t.Run("end_"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			fragment := combined[:i]
			_, _ = parse.Source(fragment)
		})
	}

	// Remove one character at each position
	for i := range combined {
		t.Run("each_"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			fragment := append([]byte(nil), combined[:i]...)
			fragment = append(fragment, combined[i+1:]...)
			_, _ = parse.Source(fragment)
		})
	}
}
