package builtin

import (
	"strings"
	"unicode"

	"github.com/vietmpl/vie/value"
)

var Functions = map[string]value.Function{
	"upper": {
		Name:     "upper",
		ArgTypes: []value.Type{value.TypeString},
		Impl:     upper,
	},
	"lower": {
		Name:     "lower",
		ArgTypes: []value.Type{value.TypeString},
		Impl:     lower,
	},
	"capitalize": {
		Name:     "capitalize",
		ArgTypes: []value.Type{value.TypeString},
		Impl:     capitalize,
	},
	"title": {
		Name:     "title",
		ArgTypes: []value.Type{value.TypeString},
		Impl:     title,
	},
	"first": {
		Name:     "first",
		ArgTypes: []value.Type{value.TypeString},
		Impl:     first,
	},
	"last": {
		Name:     "last",
		ArgTypes: []value.Type{value.TypeString},
		Impl:     last,
	},
	"reverse": {
		Name:     "reverse",
		ArgTypes: []value.Type{value.TypeString},
		Impl:     reverse,
	},
	"trimSpace": {
		Name:     "trimSpace",
		ArgTypes: []value.Type{value.TypeString},
		Impl:     trimSpace,
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
			inWord = false // Non-letter / digit
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
