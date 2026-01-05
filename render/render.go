package render

import (
	"fmt"
	"io"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/value"
)

// MustRenderTemplate renders a parsed template using the provided context.
//
// It assumes that both the template and the context have been type-checked using
// the analysis package. If either is invalid, behavior is undefined and may
// panic. For user-friendly error messages, use RenderFile instead.
func MustRenderTemplate(w io.Writer, template *ast.Template, context map[string]value.Value) {
	r := renderer{
		c: context,
		w: w,
	}
	r.renderBlocks(template.Blocks)
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
		_, _ = io.WriteString(r.w, b.Content)

	case *ast.CommentBlock:
		// Comments do not produce output.

	case *ast.DisplayBlock:
		x := r.evalExpr(b.Value)
		xv := x.(value.String)
		_, _ = io.WriteString(r.w, string(xv))

	case *ast.IfBlock:
		for _, branch := range b.Branches {
			conditionVal := r.evalExpr(branch.Condition)
			// TODO(skewb1k): fixme.
			if conditionVal.(value.Bool) {
				r.renderBlocks(branch.Consequence)
				break
			}
		}
		if b.ElseConsequence != nil {
			r.renderBlocks(*b.ElseConsequence)
		}

	default:
		panic(fmt.Sprintf("render: unexpected block type %T", block))
	}
}
