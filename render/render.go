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
	r.renderBlocks(file.Blocks)
}

type renderer struct {
	c map[string]value.Value
	w io.Writer
}

func (r *renderer) renderBlocks(blocks []ast.Block) {
	for _, b := range blocks {
		r.renderBlock(b)
	}
}

func (r *renderer) renderBlock(block ast.Block) {
	switch b := block.(type) {
	case *ast.TextBlock:
		_, _ = io.WriteString(r.w, b.Value)

	case *ast.CommentBlock:
		// Comments do not produce output.

	case *ast.RenderBlock:
		x := r.evalExpr(b.X)
		xv := x.(value.String)
		_, _ = io.WriteString(r.w, string(xv))

	case *ast.IfBlock:
		condVal := r.evalExpr(b.Cond)
		cond := condVal.(value.Bool)

		if cond {
			r.renderBlocks(b.Cons)
		} else {
			for _, elseIfClause := range b.ElseIfs {
				elseCondVal := r.evalExpr(elseIfClause.Cond)
				elseCond := elseCondVal.(value.Bool)
				if elseCond {
					r.renderBlocks(elseIfClause.Cons)
					break
				}
			}
			if b.Else != nil {
				r.renderBlocks(b.Else.Cons)
			}
		}

	case *ast.SwitchBlock:
		val := r.evalExpr(b.Value)

		for _, c := range b.Cases {
			for _, e := range c.List {
				x := r.evalExpr(e)
				if value.Eq(x, val) {
					r.renderBlocks(c.Body)
					return
				}
			}
		}
		panic("unreachable")

	default:
		panic(fmt.Sprintf("render: unexpected block type %T", block))
	}
}
