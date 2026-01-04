package value

import (
	"fmt"
	"strconv"

	"github.com/vietmpl/vie/ast"
)

type Value interface {
	Type() Type
}

type String string

func (String) Type() Type { return TypeString }

type Bool bool

func (Bool) Type() Type { return TypeBool }

func FromBasicLit(l *ast.BasicLiteral) Value {
	switch l.Kind {
	case ast.KindBool:
		return Bool(l.Value == "true")
	case ast.KindString:
		// TODO(skewb1k): Replace this hack with a manual parser.
		value, _ := strconv.Unquote(l.Value)
		return String(value)
	default:
		panic("value: unsupported literal kind")
	}
}

func (x String) Eq(y String) Bool       { return x == y }
func (x String) Neq(y String) Bool      { return x != y }
func (x String) Concat(y String) String { return x + y }

func (x Bool) Eq(y Bool) Bool  { return x == y }
func (x Bool) Neq(y Bool) Bool { return x != y }

func (x Bool) And(y Bool) Bool { return x && y }
func (x Bool) Or(y Bool) Bool  { return x || y }

func Eq[T comparable](x, y T) Bool {
	return Bool(x == y)
}

func Neq[T comparable](x, y T) Bool {
	return Bool(x != y)
}

type Function struct {
	Name       string
	ArgTypes   []Type
	ReturnType Type
	Impl       func(args []Value) Value
}

func (Function) Type() Type { return TypeFunction }

func (f *Function) Call(args []Value) (Value, error) {
	if len(args) != len(f.ArgTypes) {
		return nil, fmt.Errorf("function %s expects %d arguments, got %d", f.Name, len(f.ArgTypes), len(args))
	}
	for i, arg := range args {
		if arg.Type() != f.ArgTypes[i] {
			return nil, fmt.Errorf("argument %d: expected %v, got %v", i, f.ArgTypes[i], arg.Type())
		}
	}
	return f.Impl(args), nil
}
