package render

import (
	"fmt"
	"strings"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/builtin"
	"github.com/vietmpl/vie/value"
)

// RenderFileUnsafe renders a parsed template file using the provided context.
//
// It assumes that both the file and the context have been type-checked using
// the analysis package. If either is invalid, behavior is undefined and may
// panic. For user-friendly error messages, use RenderFile instead.
func RenderFileUnsafe(file *ast.File, context map[string]value.Value) []byte {
	r := renderer{
		c: context,
	}
	r.renderStmts(file.Stmts)
	return []byte(r.out.String())
}

type renderer struct {
	c   map[string]value.Value
	out strings.Builder
}

func (r *renderer) renderStmts(stmts []ast.Stmt) {
	for _, s := range stmts {
		r.renderStmt(s)
	}
}

func (r *renderer) renderStmt(s ast.Stmt) {
	switch n := s.(type) {
	case *ast.Text:
		r.out.WriteString(n.Value)

	case *ast.RenderStmt:
		x := r.evalExpr(n.X)
		xv := x.(value.String)
		r.out.WriteString(string(xv))

	case *ast.IfStmt:
		condVal := r.evalExpr(n.Cond)
		cond := condVal.(value.Bool)

		if cond {
			r.renderStmts(n.Cons)
		} else {
			for _, elseIfClause := range n.ElseIfs {
				elseCondVal := r.evalExpr(elseIfClause.Cond)
				elseCond := elseCondVal.(value.Bool)
				if elseCond {
					r.renderStmts(elseIfClause.Cons)
					break
				}
			}
			if n.Else != nil {
				r.renderStmts(n.Else.Cons)
			}
		}

	case *ast.SwitchStmt:
		val := r.evalExpr(n.Value)

		for _, c := range n.Cases {
			for _, e := range c.List {
				x := r.evalExpr(e)
				if value.Eq(x, val) {
					r.renderStmts(c.Body)
					return
				}
			}
		}
		panic("unreachable")

	default:
		panic(fmt.Sprintf("render: unexpected stmt type %T", s))
	}
}

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
	for _, arg := range exprList {
		v := r.evalExpr(arg)
		values = append(values, v)
	}
	return values
}
