package value

import (
	"bytes"
	"strconv"

	"github.com/vietmpl/vie/ast"
)

type Value interface {
	value()
}

type String string
type Bool bool

func (String) value() {}
func (Bool) value()   {}

func FromBasicLit(l *ast.BasicLit) Value {
	switch l.Kind {
	case ast.KindBool:
		return Bool(bytes.Equal(l.Value, []byte("true")))
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
func (x String) Gtr(y String) Bool      { return x > y }
func (x String) Geq(y String) Bool      { return x >= y }
func (x String) Lss(y String) Bool      { return x < y }
func (x String) Leq(y String) Bool      { return x <= y }
func (x String) Concat(y String) String { return x + y }

func (x Bool) Eq(y Bool) Bool  { return x == y }
func (x Bool) Neq(y Bool) Bool { return x != y }

func (x Bool) toInt() int {
	if x {
		return 1
	}
	return 0
}
func (x Bool) Gtr(y Bool) Bool { return x.toInt() > y.toInt() }
func (x Bool) Geq(y Bool) Bool { return x.toInt() >= y.toInt() }
func (x Bool) Lss(y Bool) Bool { return x.toInt() < y.toInt() }
func (x Bool) Leq(y Bool) Bool { return x.toInt() <= y.toInt() }

func (x Bool) And(y Bool) Bool { return x && y }
func (x Bool) Or(y Bool) Bool  { return x || y }

func Eq[T comparable](x, y T) Bool {
	return Bool(x == y)
}

func Neq[T comparable](x, y T) Bool {
	return Bool(x != y)
}
