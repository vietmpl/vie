package format

import (
	"fmt"
	"io"

	"github.com/vietmpl/vie/ast"
)

type formatter struct {
	w io.Writer
}

func (f *formatter) blocks(blocks []ast.Block) {
	for _, b := range blocks {
		f.block(b)
	}
}

func (f *formatter) block(b ast.Block) {
	switch n := b.(type) {
	case *ast.TextBlock:
		io.WriteString(f.w, n.Value)

	case *ast.CommentBlock:
		// TODO(skewb1k): format leading/trailing whitespaces.
		io.WriteString(f.w, "{#")
		io.WriteString(f.w, n.Content)
		io.WriteString(f.w, "#}")

	case *ast.RenderBlock:
		io.WriteString(f.w, "{{ ")
		f.expr(n.X)
		io.WriteString(f.w, " }}")

	case *ast.IfBlock:
		io.WriteString(f.w, "{% if ")
		f.expr(n.Cond)
		io.WriteString(f.w, " %}\n")
		f.blocks(n.Cons)
		for _, elseIfClause := range n.ElseIfs {
			io.WriteString(f.w, "{% else if ")
			f.expr(elseIfClause.Cond)
			io.WriteString(f.w, " %}\n")
			f.blocks(elseIfClause.Cons)
		}
		if n.Else != nil {
			io.WriteString(f.w, "{% else %}\n")
			f.blocks(n.Else.Cons)
		}
		io.WriteString(f.w, "{% end %}\n")

	case *ast.SwitchBlock:
		io.WriteString(f.w, "{% switch ")
		f.expr(n.Value)
		io.WriteString(f.w, " %}\n")
		for _, c := range n.Cases {
			io.WriteString(f.w, "{% case ")
			f.exprList(c.List)
			io.WriteString(f.w, " %}\n")
			f.blocks(c.Body)
		}
		io.WriteString(f.w, "{% end %}\n")

	default:
		panic(fmt.Sprintf("format: unexpected block type %T", b))
	}
}

func (f *formatter) expr(e ast.Expr) {
	switch n := e.(type) {
	case *ast.BasicLit:
		io.WriteString(f.w, n.Value)

	case *ast.Ident:
		io.WriteString(f.w, n.Name)

	case *ast.UnaryExpr:
		io.WriteString(f.w, n.Op.String())
		// do not insert whitespace after '!'
		if n.Op != ast.UnOpKindExcl {
			io.WriteString(f.w, " ")
		}
		f.expr(n.X)

	case *ast.BinaryExpr:
		f.expr(n.X)
		io.WriteString(f.w, " ")
		io.WriteString(f.w, n.Op.String())
		io.WriteString(f.w, " ")
		f.expr(n.Y)

	case *ast.ParenExpr:
		io.WriteString(f.w, "(")
		f.expr(n.X)
		io.WriteString(f.w, ")")

	case *ast.CallExpr:
		f.expr(&n.Func)
		io.WriteString(f.w, "(")
		f.exprList(n.Args)
		io.WriteString(f.w, ")")

	case *ast.PipeExpr:
		f.expr(n.Arg)
		io.WriteString(f.w, " | ")
		io.WriteString(f.w, n.Func.Name)

	default:
		panic(fmt.Sprintf("format: unexpected expr type %T", e))
	}
}

func (f *formatter) exprList(l []ast.Expr) {
	if len(l) > 0 {
		f.expr(l[0])
	}
	for i := 1; i < len(l); i++ {
		io.WriteString(f.w, ", ")
		f.expr(l[i])
	}
}
