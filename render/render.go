package render

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/builtin"
	"github.com/vietmpl/vie/value"
)

func Source(src *ast.SourceFile, context map[string]value.Value) ([]byte, error) {
	r := renderer{
		c: context,
	}
	if err := r.renderStmts(src.Stmts); err != nil {
		return nil, err
	}
	return r.out.Bytes(), nil
}

type renderer struct {
	c   map[string]value.Value
	out bytes.Buffer

	afterTag bool
}

func (r *renderer) renderStmts(stmts []ast.Stmt) error {
	for _, s := range stmts {
		if err := r.renderStmt(s); err != nil {
			return err
		}
	}
	return nil
}

func (r *renderer) renderStmt(s ast.Stmt) error {
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
		x, err := r.evalExpr(n.X)
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
		condVal, err := r.evalExpr(n.Cond)
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
			if err := r.renderStmts(n.Cons); err != nil {
				return err
			}
		} else {
			var elseCond value.Bool
			for _, elseIfClause := range n.ElseIfs {
				elseCondVal, err := r.evalExpr(elseIfClause.Cond)
				if err != nil {
					return err
				}
				elseCond, ok = elseCondVal.(value.Bool)
				if !ok {
					return fmt.Errorf("unexpected type in else if condition: %T", condVal)
				}
				if elseCond {
					r.truncateTrailspaces()
					if err := r.renderStmts(elseIfClause.Cons); err != nil {
						return err
					}
					break
				}
			}
			if n.Else != nil {
				r.truncateTrailspaces()
				if err := r.renderStmts(n.Else.Cons); err != nil {
					return err
				}
			}
		}
		r.truncateTrailspaces()
		r.afterTag = true
		return nil

	case *ast.SwitchStmt:
		valValue, err := r.evalExpr(n.Value)
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
				x, err := r.evalExpr(e)
				if err != nil {
					return err
				}
				if value.Eq(x, valValue) {
					r.truncateTrailspaces()
					if err := r.renderStmts(c.Body); err != nil {
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

func (r *renderer) evalExpr(expr ast.Expr) (value.Value, error) {
	switch n := expr.(type) {
	case *ast.BasicLit:
		return value.FromBasicLit(n), nil

	case *ast.Ident:
		v, exists := r.c[string(n.Name)]
		if !exists {
			return nil, fmt.Errorf("%s is undefined", n.Name)
		}
		return v, nil

	case *ast.UnaryExpr:
		x, err := r.evalExpr(n.X)
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
		x, err := r.evalExpr(n.X)
		if err != nil {
			return nil, err
		}

		y, err := r.evalExpr(n.Y)
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
		return r.evalExpr(n.X)

	case *ast.CallExpr:
		fn, err := lookupFunction(n.Func.Name)
		if err != nil {
			return nil, err
		}

		args, err := r.evalExprList(n.Args)
		if err != nil {
			return nil, err
		}
		return fn.Call(args)

	case *ast.PipeExpr:
		fn, err := lookupFunction(n.Func.Name)
		if err != nil {
			return nil, err
		}

		arg, err := r.evalExpr(n.Arg)
		if err != nil {
			return nil, err
		}
		return fn.Call([]value.Value{arg})

	default:
		panic(fmt.Sprintf("render: unexpected expr type %T", expr))
	}
}

func (r *renderer) evalExprList(exprList []ast.Expr) ([]value.Value, error) {
	vals := make([]value.Value, 0, len(exprList))
	for _, arg := range exprList {
		v, err := r.evalExpr(arg)
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

// TODO(skewb1k): rename and reconsider.
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
