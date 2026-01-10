package render

import (
	"fmt"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/builtin"
	"github.com/vietmpl/vie/token"
	"github.com/vietmpl/vie/value"
)

func (r renderer) evalExpr(expr ast.Expr) value.Value {
	switch e := expr.(type) {
	case *ast.BasicLiteral:
		return value.FromBasicLit(e)

	case *ast.Identifier:
		return r.c[e.Value]

	case *ast.UnaryExpr:
		x := r.evalExpr(e.Operand)
		// Assume that n.Op is valid and contains only '!' and 'not' operators.
		xx := x.(value.Bool)
		return !xx

	case *ast.BinaryExpr:
		x := r.evalExpr(e.LOperand)
		y := r.evalExpr(e.ROperand)
		switch e.Operator {
		case token.TILDE:
			xx := x.(value.String)
			yy := y.(value.String)
			return xx.Concat(yy)

		case token.EQUAL_EQUAL:
			return value.Eq(x, y)

		case token.BANG_EQUAL:
			return value.Neq(x, y)

		case token.KEYWORD_AND:
			// TODO: validate values.
			xx := x.(value.Bool)
			yy := y.(value.Bool)
			return xx.And(yy)

		case token.KEYWORD_OR:
			// TODO: validate values.
			xx := x.(value.Bool)
			yy := y.(value.Bool)
			return xx.Or(yy)

		default:
			panic(fmt.Sprintf("render: unexpected binary operator: %T", e.Operator))
		}

	case *ast.ParenExpr:
		return r.evalExpr(e.Value)

	case *ast.CallExpr:
		fn, _ := builtin.LookupFunction(e.Function)
		args := r.evalExprList(e.Arguments)
		res, _ := fn.Call(args)
		return res

	case *ast.PipeExpr:
		fn, _ := builtin.LookupFunction(e.Function)
		arg := r.evalExpr(e.Argument)
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
