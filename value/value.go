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

func FromBasicLit(l *ast.BasicLit) Value {
	switch l.Kind {
	case ast.KindBool:
		return Bool(l.Value == "true")
	case ast.KindString:
		// TODO(skewb1k): Replace this hack with a manual parser that handles
		// escape sequences and quote types. Currently we wrap single-quote
		// literals in quotes and use [strconv.Unquote] to parse them. This works
		// for basic cases but incorrectly allows backticks and mixes rune vs
		// string parsing.
		s := string(l.Value)
		if l.Value[0] == '\'' {
			s = "\"" + s + "\""
			v, _ := strconv.Unquote(s)
			return String(v[1 : len(v)-1])
		}
		v, _ := strconv.Unquote(s)
		return String(v)
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
