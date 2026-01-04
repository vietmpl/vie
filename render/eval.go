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
		case ast.CONCAT:
			xx := x.(value.String)
			yy := y.(value.String)
			return xx.Concat(yy)

		case ast.EQUAL:
			return value.Eq(x, y)

		case ast.NOT_EQUAL:
			return value.Neq(x, y)

		case ast.AND:
			// TODO: validate values.
			xx := x.(value.Bool)
			yy := y.(value.Bool)
			return xx.And(yy)

		case ast.OR:
			// TODO: validate values.
			xx := x.(value.Bool)
			yy := y.(value.Bool)
			return xx.Or(yy)

		default:
			panic(fmt.Sprintf("render: unexpected binary operator: %T", e.Op))
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
