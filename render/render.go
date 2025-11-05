package render

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/builtin"
	"github.com/vietmpl/vie/value"
)

type renderer struct {
	c   map[string]value.Value
	out bytes.Buffer

	afterTag bool
}

func Source(src *ast.SourceFile, context map[string]value.Value) ([]byte, error) {
	r := renderer{
		c: context,
	}
	if err := r.stmts(src.Stmts); err != nil {
		return nil, err
	}
	return r.out.Bytes(), nil
}

func (r *renderer) stmts(stmts []ast.Stmt) error {
	for _, s := range stmts {
		if err := r.stmt(s); err != nil {
			return err
		}
	}
	return nil
}

func (r *renderer) stmt(s ast.Stmt) error {
	switch n := s.(type) {
	case *ast.Text:
		if r.afterTag {
			if i := strings.IndexByte(n.Value, '\n'); i != -1 {
				n.Value = n.Value[i+1:]
			}
			r.afterTag = false
		}
		r.out.WriteString(n.Value)
		return nil

	case *ast.RenderStmt:
		x, err := evalExpr(r.c, n.X)
		if err != nil {
			return err
		}
		switch xv := x.(type) {
		case value.String:
			r.out.WriteString(string(xv))
			return nil
		default:
			panic(fmt.Sprintf("render: unexpected value type in render statement %T", x))
		}

	case *ast.IfStmt:
		condVal, err := evalExpr(r.c, n.Cond)
		if err != nil {
			return err
		}
		cond, ok := condVal.(value.Bool)
		if !ok {
			return fmt.Errorf("unexpected type in if condition: %T", condVal)
		}

		// TODO(skewb1k): refactor or document.
		r.afterTag = true

		if cond {
			r.truncateTrailspaces()
			if err := r.stmts(n.Cons); err != nil {
				return err
			}
		} else {
			var elseCond value.Bool
			for _, elseIfClause := range n.ElseIfs {
				elseCondVal, err := evalExpr(r.c, elseIfClause.Cond)
				if err != nil {
					return err
				}
				elseCond, ok = elseCondVal.(value.Bool)
				if !ok {
					return fmt.Errorf("unexpected type in else if condition: %T", condVal)
				}
				if elseCond {
					r.truncateTrailspaces()
					if err := r.stmts(elseIfClause.Cons); err != nil {
						return err
					}
					break
				}
			}
			if n.Else != nil {
				r.truncateTrailspaces()
				if err := r.stmts(n.Else.Cons); err != nil {
					return err
				}
			}
		}
		r.truncateTrailspaces()
		r.afterTag = true
		return nil

	case *ast.SwitchStmt:
		valValue, err := evalExpr(r.c, n.Value)
		if err != nil {
			return err
		}
		_, ok := valValue.(value.String)
		if !ok {
			return fmt.Errorf("unexpected type in switch value: %T", valValue)
		}

		r.afterTag = true

		for _, c := range n.Cases {
			for _, e := range c.List {
				x, err := evalExpr(r.c, e)
				if err != nil {
					return err
				}
				if value.Eq(x, valValue) {
					r.truncateTrailspaces()
					if err := r.stmts(c.Body); err != nil {
						return err
					}
					r.truncateTrailspaces()
					r.afterTag = true
					return nil
				}
			}
		}
		panic("unreachable")

	default:
		panic(fmt.Sprintf("render: unexpected stmt type %T", s))
	}
}

func evalExpr(c map[string]value.Value, e ast.Expr) (value.Value, error) {
	switch n := e.(type) {
	case *ast.BasicLit:
		return value.FromBasicLit(n), nil

	case *ast.Ident:
		v, exists := c[string(n.Name)]
		if !exists {
			return nil, fmt.Errorf("%s is undefined", n.Name)
		}
		return v, nil

	case *ast.UnaryExpr:
		x, err := evalExpr(c, n.X)
		if err != nil {
			return nil, err
		}

		switch n.Op {
		case ast.UnOpKindExcl, ast.UnOpKindNot:
			xs, ok := x.(value.Bool)
			if !ok {
				return nil, fmt.Errorf("unexpected type: %T", x)
			}
			return !xs, nil

		default:
			panic(fmt.Sprintf("render: unexpected UnOpKind: %T", n.Op))
		}

	case *ast.BinaryExpr:
		x, err := evalExpr(c, n.X)
		if err != nil {
			return nil, err
		}

		y, err := evalExpr(c, n.Y)
		if err != nil {
			return nil, err
		}
		switch n.Op {
		case ast.BinOpKindConcat:
			// TODO: improve error messages.
			xs, ok := x.(value.String)
			if !ok {
				return nil, fmt.Errorf("unexpected type in concat: %T", x)
			}
			ys, ok := y.(value.String)
			if !ok {
				return nil, fmt.Errorf("unexpected type in concat: %T", x)
			}
			return xs.Concat(ys), nil

		case ast.BinOpKindEq,
			ast.BinOpKindIs:
			return value.Eq(x, y), nil

		case ast.BinOpKindNeq,
			ast.BinOpKindIsNot:
			return value.Neq(x, y), nil

		case ast.BinOpKindGtr,
			ast.BinOpKindGeq,
			ast.BinOpKindLss,
			ast.BinOpKindLeq,
			ast.BinOpKindLAnd,
			ast.BinOpKindLOr,
			ast.BinOpKindAnd,
			ast.BinOpKindOr:

			// TODO: improve error messages.
			xs, ok := x.(value.Bool)
			if !ok {
				return nil, fmt.Errorf("unexpected type: %T", x)
			}
			ys, ok := y.(value.Bool)
			if !ok {
				return nil, fmt.Errorf("unexpected type: %T", x)
			}

			switch n.Op {
			case ast.BinOpKindGtr:
				return xs.Gtr(ys), nil

			case ast.BinOpKindGeq:
				return xs.Geq(ys), nil

			case ast.BinOpKindLss:
				return xs.Lss(ys), nil

			case ast.BinOpKindLeq:
				return xs.Leq(ys), nil

			case ast.BinOpKindLAnd,
				ast.BinOpKindAnd:
				return xs.And(ys), nil

			case ast.BinOpKindLOr,
				ast.BinOpKindOr:
				return xs.Or(ys), nil

			default:
				panic("unreachable")
			}
		default:
			panic(fmt.Sprintf("render: unexpected BinOpKind: %T", n.Op))
		}

	case *ast.ParenExpr:
		return evalExpr(c, n.X)

	case *ast.CallExpr:
		fn, err := lookupFunction(n.Func.Name)
		if err != nil {
			return nil, err
		}

		args, err := evalExprList(c, n.Args)
		if err != nil {
			return nil, err
		}
		return fn.Call(args)

	case *ast.PipeExpr:
		fn, err := lookupFunction(n.Func.Name)
		if err != nil {
			return nil, err
		}

		arg, err := evalExpr(c, n.Arg)
		if err != nil {
			return nil, err
		}
		return fn.Call([]value.Value{arg})

	default:
		panic(fmt.Sprintf("render: unexpected expr type %T", e))
	}
}

func evalExprList(c map[string]value.Value, l []ast.Expr) ([]value.Value, error) {
	vals := make([]value.Value, 0, len(l))
	for _, arg := range l {
		v, err := evalExpr(c, arg)
		if err != nil {
			return nil, err
		}
		vals = append(vals, v)
	}
	return vals, nil
}

func lookupFunction(name string) (value.Function, error) {
	if name[0] != '@' {
		return value.Function{}, fmt.Errorf(
			"function %s is undefined. Only builtin functions (starting with '@') are supported, user-defined functions are not yet implemented",
			name,
		)
	}
	fn, exists := builtin.Functions[name[1:]]
	if !exists {
		return value.Function{}, fmt.Errorf("function %s is undefined", name)
	}
	return fn, nil
}

func (r *renderer) truncateTrailspaces() {
	data := r.out.Bytes()
	i := len(data) - 1
	for i >= 0 && (data[i] == ' ' || data[i] == '\t') {
		i--
	}
	// If we reached start of buffer or the last non-space is a newline, truncate spaces
	if i < 0 || data[i] == '\n' {
		r.out.Truncate(i + 1)
	}
}
