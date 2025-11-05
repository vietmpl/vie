package builtin

import (
	"strings"

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
}

func upper(args []value.Value) value.Value {
	s := args[0].(value.String)
	return value.String(strings.ToUpper(string(s)))
}

func lower(args []value.Value) value.Value {
	s := args[0].(value.String)
	return value.String(strings.ToLower(string(s)))
}
