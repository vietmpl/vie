package render_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vietmpl/vie/parser"
	"github.com/vietmpl/vie/render"
	"github.com/vietmpl/vie/value"
)

func TestSource(t *testing.T) {
	type tc struct {
		name     string
		input    string
		expected string
		context  map[string]value.Value
	}

	tests := []tc{
		// Render
		{
			name:     "empty string",
			input:    `{{ "" }}`,
			expected: ``,
		},
		{
			name:     "double-quote string",
			input:    `{{ "Hello, world!" }}`,
			expected: `Hello, world!`,
		},
		{
			name:     "numeric string",
			input:    `{{ "123" }}`,
			expected: `123`,
		},
		{
			name:     "single-quote string",
			input:    `{{ '123' }}`,
			expected: `123`,
		},
		{
			name:     "string with newline",
			input:    `{{ "1\n23" }}`,
			expected: "1\n23",
		},
		{
			name:     "string with escaped quote",
			input:    `{{ "He said \"hi\"" }}`,
			expected: `He said "hi"`,
		},
		{
			name:     "string with escaped backslash",
			input:    `{{ "C:\\i\\love\\windows" }}`,
			expected: `C:\i\love\windows`,
		},
		{
			name:     "concatenation",
			input:    `{{ "a" ~ "b" }}`,
			expected: `ab`,
		},
		{
			name:     "variable lookup",
			input:    `{{ a }}`,
			expected: `123`,
			context: map[string]value.Value{
				"a": value.String("123"),
			},
		},
		{
			name:     "concatenation with variable",
			input:    `{{ "Value: " ~ val }}`,
			expected: `Value: 42`,
			context: map[string]value.Value{
				"val": value.String("42"),
			},
		},

		// Functions
		// TODO:
		// {
		// 	name:   "upper",
		// 	input:  `{{ upper("up") }}`,
		// 	output: `UP`,
		// },

		// If
		{
			name:     "if true",
			input:    "{% if true %}1{% end %}",
			expected: "1",
		},
		{
			name:     "if with render",
			input:    "{% if true %}{{ '123' }}{% end %}",
			expected: "123",
		},
		{
			name:     "if with render",
			input:    "{% if true %} a {{ '123' }} b {% end %}",
			expected: " a 123 b ",
		},
		{
			name:     "multiline if",
			input:    "{% if true %}\n1\n{% end %}",
			expected: "1\n",
		},
		{
			name:     "multiline if with render",
			input:    "{% if true %}\n{{ 'str' }} text \n{% end %}",
			expected: "str text \n",
		},
		{
			name:     "multiline if2",
			input:    "{% if true %}\n\n1\n\n{% end %}",
			expected: "\n1\n\n",
		},
		{
			name:     "multiline if3",
			input:    "{% if true %}\n1\n{% end %}\n",
			expected: "1\n",
		},
		{
			name:     "indented if",
			input:    "\t{% if true %}\n\t1\n\t{% end %}",
			expected: "\t1\n",
		},
		{
			name:     "multiline if else",
			input:    "{% if false %}\n1\n{% else %}\n2\n{% end %}",
			expected: "2\n",
		},
		{
			name:     "if false",
			input:    `{% if false %}1{% end %}`,
			expected: ``,
		},
		{
			name:     "if else",
			input:    `{% if false %}1{% else %}2{% end %}`,
			expected: `2`,
		},
		{
			name:     "else if",
			input:    `{% if false %}1{% else if true %}2{% end %}`,
			expected: `2`,
		},
		{
			name:     "mutliple else if's",
			input:    `{% if false %}1{% else if false %}2{% else if true %}3{% end %}`,
			expected: `3`,
		},
		{
			name:     "else if else",
			input:    `{% if false %}1{% else if false %}2{% else %}3{% end %}`,
			expected: `3`,
		},

		// Operators
		{
			name:     "empty strings equal",
			input:    `{% if "" == "" %}1{% end %}`,
			expected: `1`,
		},
		{
			name:     "strings equal",
			input:    `{% if "a" == "a" %}1{% end %}`,
			expected: `1`,
		},
		{
			name:     "strings not equal",
			input:    `{% if "a" != "b" %}1{% end %}`,
			expected: `1`,
		},
		{
			name:     "true's equal",
			input:    `{% if true == true %}1{% end %}`,
			expected: `1`,
		},
		{
			name:     "false's equal",
			input:    `{% if false == false %}1{% end %}`,
			expected: `1`,
		},
		{
			name:     "false != true",
			input:    `{% if false != true %}1{% end %}`,
			expected: `1`,
		},
		{
			name:     "true > false",
			input:    `{% if true > false %}1{% end %}`,
			expected: `1`,
		},
		{
			name:     "true >= false",
			input:    `{% if true >= false %}1{% end %}`,
			expected: `1`,
		},
		{
			name:     "false < true",
			input:    `{% if false < true %}1{% end %}`,
			expected: `1`,
		},
		{
			name:     "false or true",
			input:    `{% if false or true %}1{% end %}`,
			expected: `1`,
		},
		{
			name:     "true or true",
			input:    `{% if true or true %}1{% end %}`,
			expected: `1`,
		},
		{
			name:     "not false",
			input:    `{% if not false %}1{% end %}`,
			expected: `1`,
		},
		{
			name:     "!false",
			input:    `{% if !false %}1{% end %}`,
			expected: `1`,
		},
		{
			name:     "(true or false) and true",
			input:    `{% if (true or false) and true %}1{% end %}`,
			expected: `1`,
		},

		// Switch
		{
			name:     "simple switch",
			input:    `{% switch "123" %}{% case "123" %}1{% end %}`,
			expected: `1`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sf := parser.ParseFile([]byte(test.input))
			actual, err := render.Source(sf, test.context)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assert.Equal(t, test.expected, string(actual))
		})
	}
}
