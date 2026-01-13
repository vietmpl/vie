package format

import (
	"bytes"
	"fmt"

	"github.com/vietmpl/vie/ast"
)

type printer struct {
	buffer bytes.Buffer
}

func (p *printer) printBlocks(b []ast.Block) {
	for _, block := range b {
		p.printBlock(block)
	}
}

func (p *printer) printBlock(b ast.Block) {
	switch block := b.(type) {
	case *ast.TextBlock:
		p.buffer.WriteString(block.Content)

	case *ast.CommentBlock:
		// TODO(skewb1k): format leading/trailing whitespaces.
		p.buffer.WriteString("{#")
		p.buffer.WriteString(block.Content)
		p.buffer.WriteString("#}")

	case *ast.DisplayBlock:
		p.buffer.WriteString("{{ ")
		p.printExpr(block.Value)
		p.buffer.WriteString(" }}")

	case *ast.IfBlock:
		p.buffer.WriteString("{% if ")
		branch0 := block.Branches[0]
		p.printExpr(branch0.Condition)
		p.buffer.WriteString(" %}")
		p.printBlocks(branch0.Consequence)

		for _, branch := range block.Branches[1:] {
			p.buffer.WriteString("{% elseif ")
			p.printExpr(branch.Condition)
			p.buffer.WriteString(" %}")
			p.printBlocks(branch.Consequence)
		}

		if block.Alternative != nil {
			p.buffer.WriteString("{% else %}")
			p.printBlocks(*block.Alternative)
		}
		p.buffer.WriteString("{% end %}")

	default:
		panic(fmt.Sprintf("format: unexpected block type %T", b))
	}
}

func (p *printer) printExpr(e ast.Expr) {
	switch expr := e.(type) {
	case *ast.BasicLiteral:
		p.buffer.WriteString(expr.Value)

	case *ast.Identifier:
		p.buffer.WriteString(expr.Value)

	case *ast.UnaryExpr:
		p.buffer.WriteString(expr.Operator.String())
		p.printExpr(expr.Operand)

	case *ast.BinaryExpr:
		p.printExpr(expr.LOperand)
		p.buffer.WriteByte(' ')
		p.buffer.WriteString(expr.Operator.String())
		p.buffer.WriteByte(' ')
		p.printExpr(expr.ROperand)

	case *ast.ParenExpr:
		p.buffer.WriteByte('(')
		p.printExpr(expr.Value)
		p.buffer.WriteByte(')')

	case *ast.CallExpr:
		p.printExpr(&expr.Function)
		p.buffer.WriteByte('(')
		p.printExprList(expr.Arguments)
		p.buffer.WriteByte(')')

	case *ast.PipeExpr:
		p.printExpr(expr.Argument)
		p.buffer.WriteString(" | ")
		p.buffer.WriteString(expr.Function.Value)

	default:
		panic(fmt.Sprintf("format: unexpected expr type %T", e))
	}
}

func (p *printer) printExprList(el []ast.Expr) {
	if len(el) > 0 {
		p.printExpr(el[0])
	}
	for i := 1; i < len(el); i++ {
		p.buffer.WriteString(", ")
		p.printExpr(el[i])
	}
}
