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
		io.WriteString(f.w, n.Content)

	case *ast.CommentBlock:
		// TODO(skewb1k): format leading/trailing whitespaces.
		io.WriteString(f.w, "{#")
		io.WriteString(f.w, n.Content)
		io.WriteString(f.w, "#}")

	case *ast.DisplayBlock:
		io.WriteString(f.w, "{{ ")
		f.expr(n.Value)
		io.WriteString(f.w, " }}")

	case *ast.IfBlock:
		io.WriteString(f.w, "{% if ")
		f.expr(n.Condition)
		io.WriteString(f.w, " %}\n")
		f.blocks(n.Consequence)
		for _, elseIfClause := range n.ElseIfs {
			io.WriteString(f.w, "{% else if ")
			f.expr(elseIfClause.Condition)
			io.WriteString(f.w, " %}\n")
			f.blocks(elseIfClause.Consequence)
		}
		if n.Else != nil {
			io.WriteString(f.w, "{% else %}\n")
			f.blocks(n.Else.Consequence)
		}
		io.WriteString(f.w, "{% end %}\n")

	default:
		panic(fmt.Sprintf("format: unexpected block type %T", b))
	}
}

func (f *formatter) expr(e ast.Expr) {
	switch n := e.(type) {
	case *ast.BasicLiteral:
		io.WriteString(f.w, n.Value)

	case *ast.Identifier:
		io.WriteString(f.w, n.Value)

	case *ast.UnaryExpr:
		io.WriteString(f.w, n.Operator.String())
		f.expr(n.Operand)

	case *ast.BinaryExpr:
		f.expr(n.LOperand)
		io.WriteString(f.w, " ")
		io.WriteString(f.w, n.Operator.String())
		io.WriteString(f.w, " ")
		f.expr(n.ROperand)

	case *ast.ParenExpr:
		io.WriteString(f.w, "(")
		f.expr(n.Value)
		io.WriteString(f.w, ")")

	case *ast.CallExpr:
		f.expr(&n.Function)
		io.WriteString(f.w, "(")
		f.exprList(n.Arguments)
		io.WriteString(f.w, ")")

	case *ast.PipeExpr:
		f.expr(n.Argument)
		io.WriteString(f.w, " | ")
		io.WriteString(f.w, n.Function.Value)

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
