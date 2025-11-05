package format

import (
	"bytes"
	"fmt"

	"github.com/vietmpl/vie/ast"
)

type formatter struct {
	out   bytes.Buffer
	level uint
}

func Source(src *ast.SourceFile) []byte {
	f := formatter{
		// TODO(skewb1k): consider pre-alloc.
		out:   bytes.Buffer{},
		level: 0,
	}
	f.stmts(src.Stmts)
	return f.out.Bytes()
}

func (f *formatter) stmts(stmts []ast.Stmt) {
	for _, s := range stmts {
		f.stmt(s)
	}
}

func (f *formatter) stmt(s ast.Stmt) {
	switch n := s.(type) {
	case *ast.Text:
		f.out.WriteString(n.Value)

	case *ast.RenderStmt:
		f.out.WriteString("{{ ")
		f.expr(n.X)
		f.out.WriteString(" }}")

	case *ast.IfStmt:
		f.out.WriteString("{% if ")
		f.expr(n.Cond)
		f.out.WriteString(" %}")
		f.stmts(n.Cons)
		for _, elseIfClause := range n.ElseIfs {
			f.out.WriteString("{% else if ")
			f.expr(elseIfClause.Cond)
			f.out.WriteString(" %}")
			f.stmts(elseIfClause.Cons)
		}
		if n.Else != nil {
			f.out.WriteString("{% else %}")
			f.stmts(n.Else.Cons)
		}
		f.out.WriteString("{% end %}")

	case *ast.SwitchStmt:
		f.out.WriteString("{% switch ")
		f.expr(n.Value)
		f.out.WriteString(" %}\n")
		for _, c := range n.Cases {
			f.out.WriteString("{% case ")
			f.exprList(c.List)
			f.out.WriteString(" %}")
			f.stmts(c.Body)
		}
		f.out.WriteString("{% end %}")

	default:
		panic(fmt.Sprintf("format: unexpected stmt type %T", s))
	}
}

func (f *formatter) expr(e ast.Expr) {
	switch n := e.(type) {
	case *ast.BasicLit:
		f.out.WriteString(n.Value)

	case *ast.Ident:
		f.out.WriteString(n.Name)

	case *ast.UnaryExpr:
		f.out.WriteString(n.Op.String())
		// do not insert whitespace after '!'
		if n.Op != ast.UnOpKindExcl {
			f.out.WriteByte(' ')
		}
		f.expr(n.X)

	case *ast.BinaryExpr:
		f.expr(n.X)
		f.out.WriteByte(' ')
		f.out.WriteString(n.Op.String())
		f.out.WriteByte(' ')
		f.expr(n.Y)

	case *ast.ParenExpr:
		f.out.WriteByte('(')
		f.expr(n.X)
		f.out.WriteByte(')')

	case *ast.CallExpr:
		f.expr(&n.Func)
		f.out.WriteByte('(')
		f.exprList(n.Args)
		f.out.WriteByte(')')

	case *ast.PipeExpr:
		f.expr(n.Arg)
		f.out.WriteString(" | ")
		f.out.WriteString(n.Func.Name)

	default:
		panic(fmt.Sprintf("format: unexpected expr type %T", e))
	}
}

func (f *formatter) exprList(l []ast.Expr) {
	if len(l) > 0 {
		f.expr(l[0])
	}
	for i := 1; i < len(l); i++ {
		f.out.WriteString(", ")
		f.expr(l[i])
	}
}
