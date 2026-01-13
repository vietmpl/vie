package format

import (
	"bytes"
	"fmt"

	"github.com/vietmpl/vie/ast"
)

type formatter struct {
	buffer bytes.Buffer
}

func (f *formatter) blocks(blocks []ast.Block) {
	for _, b := range blocks {
		f.block(b)
	}
}

func (f *formatter) block(b ast.Block) {
	switch n := b.(type) {
	case *ast.TextBlock:
		f.buffer.WriteString(n.Content)

	case *ast.CommentBlock:
		// TODO(skewb1k): format leading/trailing whitespaces.
		f.buffer.WriteString("{#")
		f.buffer.WriteString(n.Content)
		f.buffer.WriteString("#}")

	case *ast.DisplayBlock:
		f.buffer.WriteString("{{ ")
		f.expr(n.Value)
		f.buffer.WriteString(" }}")

	case *ast.IfBlock:
		f.buffer.WriteString("{% if ")
		branch0 := n.Branches[0]
		f.expr(branch0.Condition)
		f.buffer.WriteString(" %}")
		f.blocks(branch0.Consequence)
		for _, branch := range n.Branches[1:] {
			f.buffer.WriteString("{% elseif ")
			f.expr(branch.Condition)
			f.buffer.WriteString(" %}")
			f.blocks(branch.Consequence)
		}
		if n.Alternative != nil {
			f.buffer.WriteString("{% else %}")
			f.blocks(*n.Alternative)
		}
		f.buffer.WriteString("{% end %}")

	default:
		panic(fmt.Sprintf("format: unexpected block type %T", b))
	}
}

func (f *formatter) expr(e ast.Expr) {
	switch n := e.(type) {
	case *ast.BasicLiteral:
		f.buffer.WriteString(n.Value)

	case *ast.Identifier:
		f.buffer.WriteString(n.Value)

	case *ast.UnaryExpr:
		f.buffer.WriteString(n.Operator.String())
		f.expr(n.Operand)

	case *ast.BinaryExpr:
		f.expr(n.LOperand)
		f.buffer.WriteByte(' ')
		f.buffer.WriteString(n.Operator.String())
		f.buffer.WriteByte(' ')
		f.expr(n.ROperand)

	case *ast.ParenExpr:
		f.buffer.WriteByte('(')
		f.expr(n.Value)
		f.buffer.WriteByte(')')

	case *ast.CallExpr:
		f.expr(&n.Function)
		f.buffer.WriteByte('(')
		f.exprList(n.Arguments)
		f.buffer.WriteByte(')')

	case *ast.PipeExpr:
		f.expr(n.Argument)
		f.buffer.WriteString(" | ")
		f.buffer.WriteString(n.Function.Value)

	default:
		panic(fmt.Sprintf("format: unexpected expr type %T", e))
	}
}

func (f *formatter) exprList(l []ast.Expr) {
	if len(l) > 0 {
		f.expr(l[0])
	}
	for i := 1; i < len(l); i++ {
		f.buffer.WriteString(", ")
		f.expr(l[i])
	}
}
