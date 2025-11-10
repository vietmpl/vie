package render

import (
	"fmt"
	"io"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/value"
)

// MustRenderFile renders a parsed template using the provided context.
//
// It assumes that both the file and the context have been type-checked using
// the analysis package. If either is invalid, behavior is undefined and may
// panic. For user-friendly error messages, use RenderFile instead.
func MustRenderFile(w io.Writer, file *ast.File, context map[string]value.Value) {
	r := renderer{
		c: context,
		w: w,
	}
	r.renderStmts(file.Stmts)
}

type renderer struct {
	c map[string]value.Value
	w io.Writer
}

func (r *renderer) renderStmts(stmts []ast.Stmt) {
	for _, s := range stmts {
		r.renderStmt(s)
	}
}

func (r *renderer) renderStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.Text:
		_, _ = io.WriteString(r.w, s.Value)

	case *ast.RenderStmt:
		x := r.evalExpr(s.X)
		xv := x.(value.String)
		_, _ = io.WriteString(r.w, string(xv))

	case *ast.IfStmt:
		condVal := r.evalExpr(s.Cond)
		cond := condVal.(value.Bool)

		if cond {
			r.renderStmts(s.Cons)
		} else {
			for _, elseIfClause := range s.ElseIfs {
				elseCondVal := r.evalExpr(elseIfClause.Cond)
				elseCond := elseCondVal.(value.Bool)
				if elseCond {
					r.renderStmts(elseIfClause.Cons)
					break
				}
			}
			if s.Else != nil {
				r.renderStmts(s.Else.Cons)
			}
		}

	case *ast.SwitchStmt:
		val := r.evalExpr(s.Value)

		for _, c := range s.Cases {
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
		panic(fmt.Sprintf("render: unexpected stmt type %T", stmt))
	}
}
