package builtin

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/value"
)

func LookupFunction(ident ast.Ident) (value.Function, error) {
	name := ident.Name
	if name[0] != '@' {
		return value.Function{}, fmt.Errorf(
			"function %s is undefined. Only builtin functions (starting with '@') are supported, user-defined functions are not yet implemented",
			name,
		)
	}
	fn, exists := functions[name[1:]]
	if !exists {
		return value.Function{}, fmt.Errorf("function %s is undefined", name)
	}
	return fn, nil
}

var functions = map[string]value.Function{
	"upper": {
		Name:       "upper",
		ArgTypes:   []value.Type{value.TypeString},
		ReturnType: value.TypeString,
		Impl:       upper,
	},
	"lower": {
		Name:       "lower",
		ArgTypes:   []value.Type{value.TypeString},
		ReturnType: value.TypeString,
		Impl:       lower,
	},
	"capitalize": {
		Name:       "capitalize",
		ArgTypes:   []value.Type{value.TypeString},
		ReturnType: value.TypeString,
		Impl:       capitalize,
	},
	"title": {
		Name:       "title",
		ArgTypes:   []value.Type{value.TypeString},
		ReturnType: value.TypeString,
		Impl:       title,
	},
	"first": {
		Name:       "first",
		ArgTypes:   []value.Type{value.TypeString},
		ReturnType: value.TypeString,
		Impl:       first,
	},
	"last": {
		Name:       "last",
		ArgTypes:   []value.Type{value.TypeString},
		ReturnType: value.TypeString,
		Impl:       last,
	},
	"reverse": {
		Name:       "reverse",
		ArgTypes:   []value.Type{value.TypeString},
		ReturnType: value.TypeString,
		Impl:       reverse,
	},
	"trimSpace": {
		Name:       "trimSpace",
		ArgTypes:   []value.Type{value.TypeString},
		ReturnType: value.TypeString,
		Impl:       trimSpace,
	},
	"camel": {
		Name:       "camel",
		ArgTypes:   []value.Type{value.TypeString},
		ReturnType: value.TypeString,
		Impl:       camel,
	},
	"pascal": {
		Name:       "pascal",
		ArgTypes:   []value.Type{value.TypeString},
		ReturnType: value.TypeString,
		Impl:       pascal,
	},
	"kebab": {
		Name:       "kebab",
		ArgTypes:   []value.Type{value.TypeString},
		ReturnType: value.TypeString,
		Impl:       kebab,
	},
	"constant": {
		Name:       "constant",
		ArgTypes:   []value.Type{value.TypeString},
		ReturnType: value.TypeString,
		Impl:       constant,
	},
	"snake": {
		Name:       "snake",
		ArgTypes:   []value.Type{value.TypeString},
		ReturnType: value.TypeString,
		Impl:       snake,
	},
}

func upper(args []value.Value) value.Value {
	s := args[0].(value.String)
	return value.String(strings.ToUpper(string(s)))
}

func lower(args []value.Value) value.Value {
	s := args[0].(value.String)
	return value.String(strings.ToLower(string(s)))
}

func capitalize(args []value.Value) value.Value {
	s := args[0].(value.String)

	if len(s) == 0 {
		return value.String("")
	}

	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	if len(runes) > 1 {
		for i := 1; i < len(runes); i++ {
			runes[i] = unicode.ToLower(runes[i])
		}
	}
	return value.String(string(runes))
}

func title(args []value.Value) value.Value {
	s := args[0].(value.String)
	runes := []rune(s)
	inWord := false

	for i, r := range runes {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if !inWord {
				runes[i] = unicode.ToUpper(r)
				inWord = true
			} else {
				runes[i] = unicode.ToLower(r)
			}
		} else {
			inWord = false
		}
	}

	return value.String(string(runes))
}

func first(args []value.Value) value.Value {
	s := args[0].(value.String)

	if len(s) == 0 {
		return value.String("")
	}

	runes := []rune(s)
	f := runes[0]
	return value.String(string(f))
}

func last(args []value.Value) value.Value {
	s := args[0].(value.String)

	if len(s) == 0 {
		return value.String("")
	}

	runes := []rune(s)
	f := runes[len(runes)-1]
	return value.String(string(f))
}

func reverse(args []value.Value) value.Value {
	s := args[0].(value.String)
	runes := []rune(s)

	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return value.String(string(runes))
}

func trimSpace(args []value.Value) value.Value {
	s := args[0].(value.String)
	return value.String(strings.TrimSpace(string(s)))
}

func camel(args []value.Value) value.Value {
	s := args[0].(value.String)
	words := splitWords(string(s))

	if len(words) == 0 {
		return value.String("")
	}

	words[0] = strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		if len(words[i]) > 0 {
			words[i] = strings.ToUpper(string(words[i][0])) + strings.ToLower(words[i][1:])
		}
	}
	return value.String(strings.Join(words, ""))
}

func pascal(args []value.Value) value.Value {
	s := args[0].(value.String)
	words := splitWords(string(s))

	for i := range words {
		if len(words[i]) > 0 {
			words[i] = strings.ToUpper(string(words[i][0])) + strings.ToLower(words[i][1:])
		}
	}
	return value.String(strings.Join(words, ""))
}

func kebab(args []value.Value) value.Value {
	s := args[0].(value.String)
	words := splitWords(string(s))

	for i := range words {
		words[i] = strings.ToLower(words[i])
	}
	return value.String(strings.Join(words, "-"))
}

func constant(args []value.Value) value.Value {
	s := args[0].(value.String)
	words := splitWords(string(s))

	for i := range words {
		words[i] = strings.ToUpper(words[i])
	}
	return value.String(strings.Join(words, "_"))
}

func snake(args []value.Value) value.Value {
	s := args[0].(value.String)
	words := splitWords(string(s))

	for i := range words {
		words[i] = strings.ToLower(words[i])
	}
	return value.String(strings.Join(words, "_"))
}

func splitWords(s string) []string {
	var words []string
	var buf []rune

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if len(buf) > 0 && unicode.IsUpper(r) && unicode.IsLower(buf[len(buf)-1]) {
				words = append(words, string(buf))
				buf = []rune{r}
			} else {
				buf = append(buf, r)
			}
		} else if len(buf) > 0 {
			words = append(words, string(buf))
			buf = nil
		}
	}

	if len(buf) > 0 {
		words = append(words, string(buf))
	}
	return words
}
