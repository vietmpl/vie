package format

import (
	"fmt"
	"io"

	"github.com/vietmpl/vie/ast"
)

type formatter struct {
	w io.Writer
}

func FormatFile(w io.Writer, file *ast.File) error {
	f := formatter{
		w: w,
	}
	f.stmts(file.Stmts)
	return nil
}

func (f *formatter) stmts(stmts []ast.Stmt) {
	for _, s := range stmts {
		f.stmt(s)
	}
}

func (f *formatter) stmt(s ast.Stmt) {
	switch n := s.(type) {
	case *ast.Text:
		io.WriteString(f.w, n.Value)

	case *ast.RenderStmt:
		io.WriteString(f.w, "{{ ")
		f.expr(n.X)
		io.WriteString(f.w, " }}")

	case *ast.IfStmt:
		io.WriteString(f.w, "{% if ")
		f.expr(n.Cond)
		io.WriteString(f.w, " %}\n")
		f.stmts(n.Cons)
		for _, elseIfClause := range n.ElseIfs {
			io.WriteString(f.w, "{% else if ")
			f.expr(elseIfClause.Cond)
			io.WriteString(f.w, " %}\n")
			f.stmts(elseIfClause.Cons)
		}
		if n.Else != nil {
			io.WriteString(f.w, "{% else %}\n")
			f.stmts(n.Else.Cons)
		}
		io.WriteString(f.w, "{% end %}\n")

	case *ast.SwitchStmt:
		io.WriteString(f.w, "{% switch ")
		f.expr(n.Value)
		io.WriteString(f.w, " %}\n")
		for _, c := range n.Cases {
			io.WriteString(f.w, "{% case ")
			f.exprList(c.List)
			io.WriteString(f.w, " %}\n")
			f.stmts(c.Body)
		}
		io.WriteString(f.w, "{% end %}\n")

	default:
		panic(fmt.Sprintf("format: unexpected stmt type %T", s))
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
