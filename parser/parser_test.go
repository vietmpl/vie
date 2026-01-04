package parser_test

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	"github.com/vietmpl/vie/parser"
)

// TestParseBytesFuzz checks that ParseBytes never panics on incomplete
// or slightly modified input. It concatenates all files in "testdata", then
// tests truncating from the start, truncating from the end, and removing each
// character.
func TestParseBytesFuzz(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("testdata/*")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)

	var combined []byte
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		combined = append(combined, b...)
	}

	src := combined

	// Truncate from start
	for i := range src {
		t.Run("start_"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			fragment := src[i:]
			_, _ = parser.ParseBytes(fragment)
		})
	}

	// Truncate from end
	for i := len(src); i >= 0; i-- {
		t.Run("end_"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			fragment := src[:i]
			_, _ = parser.ParseBytes(fragment)
		})
	}

	// Remove one character at each position
	for i := range src {
		t.Run("each_"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			fragment := append([]byte(nil), src[:i]...)
			fragment = append(fragment, src[i+1:]...)
			_, _ = parser.ParseBytes(fragment)
		})
	}
}
