package analysis

import (
	"fmt"

	"github.com/vietmpl/vie/value"
)

// TypeVar represents an identifier whose concrete type cannot be directly
// inferred in the current context. It serves as a placeholder until all
// usages are analyzed, at which point its type is inferred from context.
type TypeVar string

func (tv TypeVar) String() string {
	return string(tv)
}

func MergeTypes(typemap map[string]value.Type, data map[string]value.Value) []error {
	var errors []error
	for varname, typ := range typemap {
		val, ok := data[varname]
		if ok {
			if val.Type() != typ {
				errors = append(errors, fmt.Errorf("%s: expected %s, got %s\n",
					varname, typ, val.Type()))
			}
		} else {
			// Assign a default value for missing variables.
			// TODO(skewb1k): maybe the parser should handle undefined variables.
			switch typ {
			case value.TypeBool:
				data[varname] = value.Bool(false)
			case value.TypeString:
				data[varname] = value.String("")
			default:
				panic(fmt.Sprintf("unexpected Type value: %d", typ))
			}
		}
	}
	return errors
}
