package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vietmpl/vie/value"
)

type testCase struct {
	input    value.String
	expected value.String
}

func runFuncTests(t *testing.T, fn func([]value.Value) value.Value, cases []testCase) {
	t.Helper()
	for _, tt := range cases {
		t.Run(string(tt.input), func(t *testing.T) {
			actual := fn([]value.Value{tt.input})
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestUpperFunc(t *testing.T) {
	tests := []testCase{
		{"", ""},
		{"hello", "HELLO"},
		{"HeLLo", "HELLO"},
	}
	runFuncTests(t, upper, tests)
}

func TestLowerFunc(t *testing.T) {
	tests := []testCase{
		{"", ""},
		{"WORLD", "world"},
		{"HeLLo", "hello"},
	}
	runFuncTests(t, lower, tests)
}
