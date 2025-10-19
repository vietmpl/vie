package format

import (
	"fmt"

	"github.com/vietmpl/vie/ast"
)

type formatter struct {
	out   []byte
	level uint
}

func newFormatter() *formatter {
	return &formatter{
		// TODO(skewb1k): consider pre-alloc.
		out:   make([]byte, 0),
		level: 0,
	}
}

func Source(src *ast.SourceFile) []byte {
	f := newFormatter()
	f.stmts(src.Stmts)
	return f.out
}

func (f *formatter) stmts(stmts []ast.Stmt) {
	for _, s := range stmts {
		f.stmt(s)
	}
}

func (f *formatter) stmt(s ast.Stmt) {
	switch n := s.(type) {
	case *ast.Text:
		f.out = append(f.out, n.Value...)

	case *ast.RenderStmt:
		f.out = append(f.out, "{{ "...)
		f.expr(n.X)
		f.out = append(f.out, " }}"...)

	case *ast.IfStmt:
		f.out = append(f.out, "{% if "...)
		f.expr(n.Cond)
		f.out = append(f.out, " %}"...)
		f.stmts(n.Cons)
		f.alt(n.Alt)
		f.out = append(f.out, "{% end %}"...)

	case *ast.SwitchStmt:
		f.out = append(f.out, "{% switch "...)
		f.expr(n.Value)
		f.out = append(f.out, " %}\n"...)
		for _, c := range n.Cases {
			f.out = append(f.out, "{% case "...)
			f.exprList(c.List)
			f.out = append(f.out, " %}"...)
			f.stmts(c.Body)
		}
		f.out = append(f.out, "{% end %}"...)

	default:
		panic(fmt.Sprintf("format: unexpected stmt type %T", s))
	}
}

func (f *formatter) alt(a any) {
	// Note: won't catch var c *ast.ElseClause = nil
	if a == nil {
		return
	}
	switch n := a.(type) {
	case *ast.ElseClause:
		f.out = append(f.out, "{% else %}"...)
		f.stmts(n.Cons)

	case *ast.ElseIfClause:
		f.out = append(f.out, "{% else if "...)
		f.expr(n.Cond)
		f.out = append(f.out, " %}"...)
		f.stmts(n.Cons)
		f.alt(n.Alt)

	default:
		panic(fmt.Sprintf("format: unexpected alternative type %T", a))
	}
}

func (f *formatter) expr(e ast.Expr) {
	switch n := e.(type) {
	case *ast.BasicLit:
		f.out = append(f.out, n.Value...)

	case *ast.Ident:
		f.out = append(f.out, n.Name...)

	case *ast.UnaryExpr:
		f.out = append(f.out, n.Op.String()...)
		// do not insert whitespace after '!'
		if n.Op != ast.UnOpKindExcl {
			f.out = append(f.out, ' ')
		}
		f.expr(n.X)

	case *ast.BinaryExpr:
		f.expr(n.X)
		f.out = append(f.out, ' ')
		f.out = append(f.out, n.Op.String()...)
		f.out = append(f.out, ' ')
		f.expr(n.Y)

	case *ast.ParenExpr:
		f.out = append(f.out, '(')
		f.expr(n.X)
		f.out = append(f.out, ')')

	case *ast.CallExpr:
		f.expr(&n.Fn)
		f.out = append(f.out, '(')
		f.exprList(n.Args)
		f.out = append(f.out, ')')

	case *ast.PipeExpr:
		f.expr(n.Arg)
		f.out = append(f.out, " | "...)
		f.out = append(f.out, n.Func.Name...)

	default:
		panic(fmt.Sprintf("format: unexpected expr type %T", e))
	}
}

func (f *formatter) exprList(l []ast.Expr) {
	if len(l) > 0 {
		f.expr(l[0])
	}
	for i := 1; i < len(l); i++ {
		f.out = append(f.out, ", "...)
		f.expr(l[i])
	}
}
