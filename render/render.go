package render

import (
	"bytes"
	"fmt"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/builtin"
	"github.com/vietmpl/vie/value"
)

func File(file *ast.File, context map[string]value.Value) ([]byte, error) {
	r := renderer{
		c: context,
	}
	if err := r.renderStmts(file.Stmts); err != nil {
		return nil, err
	}
	return r.out.Bytes(), nil
}

type renderer struct {
	c   map[string]value.Value
	out bytes.Buffer
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

		if cond {
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
					if err := r.renderStmts(elseIfClause.Cons); err != nil {
						return err
					}
					break
				}
			}
			if n.Else != nil {
				if err := r.renderStmts(n.Else.Cons); err != nil {
					return err
				}
			}
		}
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

		for _, c := range n.Cases {
			for _, e := range c.List {
				x, err := r.evalExpr(e)
				if err != nil {
					return err
				}
				if value.Eq(x, valValue) {
					if err := r.renderStmts(c.Body); err != nil {
						return err
					}
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
