package render

import (
	"fmt"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/builtin"
	"github.com/vietmpl/vie/value"
)

func (r renderer) evalExpr(expr ast.Expr) value.Value {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return value.FromBasicLit(e)

	case *ast.Ident:
		return r.c[e.Name]

	case *ast.UnaryExpr:
		x := r.evalExpr(e.X)
		// Assume that n.Op is valid and contains only '!' and 'not' operators.
		xx := x.(value.Bool)
		return !xx

	case *ast.BinaryExpr:
		x := r.evalExpr(e.X)
		y := r.evalExpr(e.Y)
		switch e.Op {
		case ast.BinOpKindConcat:
			xx := x.(value.String)
			yy := y.(value.String)
			return xx.Concat(yy)

		case ast.BinOpKindEq,
			ast.BinOpKindIs:
			return value.Eq(x, y)

		case ast.BinOpKindNeq,
			ast.BinOpKindIsNot:
			return value.Neq(x, y)

		case ast.BinOpKindGtr,
			ast.BinOpKindGeq,
			ast.BinOpKindLss,
			ast.BinOpKindLeq,
			ast.BinOpKindLAnd,
			ast.BinOpKindLOr,
			ast.BinOpKindAnd,
			ast.BinOpKindOr:

			xx := x.(value.Bool)
			yy := y.(value.Bool)

			switch e.Op {
			case ast.BinOpKindGtr:
				return xx.Gtr(yy)

			case ast.BinOpKindGeq:
				return xx.Geq(yy)

			case ast.BinOpKindLss:
				return xx.Lss(yy)

			case ast.BinOpKindLeq:
				return xx.Leq(yy)

			case ast.BinOpKindLAnd,
				ast.BinOpKindAnd:
				return xx.And(yy)

			case ast.BinOpKindLOr,
				ast.BinOpKindOr:
				return xx.Or(yy)

			default:
				panic("unreachable")
			}
		default:
			panic(fmt.Sprintf("render: unexpected BinOpKind: %T", e.Op))
		}

	case *ast.ParenExpr:
		return r.evalExpr(e.X)

	case *ast.CallExpr:
		fn, _ := builtin.LookupFunction(e.Func)
		args := r.evalExprList(e.Args)
		res, _ := fn.Call(args)
		return res

	case *ast.PipeExpr:
		fn, _ := builtin.LookupFunction(e.Func)
		arg := r.evalExpr(e.Arg)
		res, _ := fn.Call([]value.Value{arg})
		return res

	default:
		panic(fmt.Sprintf("render: unexpected expr type %T", expr))
	}
}

func (r renderer) evalExprList(exprList []ast.Expr) []value.Value {
	values := make([]value.Value, 0, len(exprList))
	for _, expr := range exprList {
		v := r.evalExpr(expr)
		values = append(values, v)
	}
	return values
}
