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
			t.Parallel()
			actual := fn([]value.Value{tt.input})
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestUpperFunc(t *testing.T) {
	t.Parallel()
	tests := []testCase{
		{"", ""},
		{"hello", "HELLO"},
		{"HeLLo", "HELLO"},
		{"hello, 世界", "HELLO, 世界"},
		{"Oly, eine echte Schönheit", "OLY, EINE ECHTE SCHÖNHEIT"},
	}
	runFuncTests(t, upper, tests)
}

func TestLowerFunc(t *testing.T) {
	t.Parallel()
	tests := []testCase{
		{"", ""},
		{"WORLD", "world"},
		{"HeLlo, 世界", "hello, 世界"},
		{"OLY, EINE ECHTE SCHÖNHEIT", "oly, eine echte schönheit"},
	}
	runFuncTests(t, lower, tests)
}

func TestCapitalizeFunc(t *testing.T) {
	t.Parallel()
	tests := []testCase{
		{"", ""},
		{"hello", "Hello"},
		{"HeLlo, 世界", "Hello, 世界"},
		{"OLY, EINE ECHTE SCHÖNHEIT", "Oly, eine echte schönheit"},
	}
	runFuncTests(t, capitalize, tests)
}

func TestTitleFunc(t *testing.T) {
	t.Parallel()
	tests := []testCase{
		{"", ""},
		{"hello", "Hello"},
		{"heLLo world", "Hello World"},
		{"hello world", "Hello World"},
		{"HELLO WORLD", "Hello World"},
		{"go-lang is FUN", "Go-Lang Is Fun"},
		{"123abc test-case", "123abc Test-Case"},
		{"mixed CAPS and lowercase", "Mixed Caps And Lowercase"},
		{"already Title Case", "Already Title Case"},
		{"ümlaut über alles", "Ümlaut Über Alles"},
		{"Oly, eine echte Schönheit", "Oly, Eine Echte Schönheit"},
	}
	runFuncTests(t, title, tests)
}

func TestFirstFunc(t *testing.T) {
	t.Parallel()
	tests := []testCase{
		{"", ""},
		{"Oly, eine echte Schönheit", "O"},
		{"世界, hello", "世"},
	}
	runFuncTests(t, first, tests)
}

func TestLastFunc(t *testing.T) {
	t.Parallel()
	tests := []testCase{
		{"", ""},
		{"2", "2"},
		{"Oly, eine echte Schönheit", "t"},
		{"Hello, 世界", "界"},
	}
	runFuncTests(t, last, tests)
}

func TestReverseFunc(t *testing.T) {
	t.Parallel()
	tests := []testCase{
		{"", ""},
		{"Oly, eine echte Schönheit", "tiehnöhcS ethce enie ,ylO"},
		{"Hello, 世界", "界世 ,olleH"},
	}
	runFuncTests(t, reverse, tests)
}

func TestTrimFunc(t *testing.T) {
	t.Parallel()
	tests := []testCase{
		{"", ""},
		{"   ", ""},
		{"\t\n\r", ""},
		{"  Hello  ", "Hello"},
		{"\tHello\n", "Hello"},
		{"  Oli, eine echte Schönheit  ", "Oli, eine echte Schönheit"},
		{"Hello, 世界", "Hello, 世界"},
	}
	runFuncTests(t, trim, tests)
}
