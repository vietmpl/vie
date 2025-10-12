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
	f.blocks(src.Blocks)
	return f.out
}

func (f *formatter) blocks(blocks []ast.Block) {
	for _, b := range blocks {
		f.block(b)
	}
}

func (f *formatter) block(b ast.Block) {
	switch n := b.(type) {
	case *ast.TextBlock:
		f.out = append(f.out, n.Value...)
	case *ast.RenderBlock:
		f.out = append(f.out, "{{ "...)
		f.expr(n.Expr)
		f.out = append(f.out, " }}"...)
	default:
		panic(fmt.Sprintf("format: unexpected block kind %T", b))
	}
}

func (f *formatter) expr(e ast.Expr) {
	switch n := e.(type) {
	case *ast.BasicLit:
		f.out = append(f.out, n.Value...)

	case *ast.Ident:
		f.out = append(f.out, n.Value...)

	case *ast.UnaryExpr:
		f.out = append(f.out, n.Op...)

		// Do not insert whitespace after '!'
		if n.Op[0] != '!' {
			f.out = append(f.out, ' ')
		}
		f.expr(n.Expr)

	case *ast.BinaryExpr:
		f.expr(n.Left)
		f.out = append(f.out, ' ')
		f.out = append(f.out, n.Op...)
		f.out = append(f.out, ' ')
		f.expr(n.Right)

	case *ast.ParenExpr:
		f.out = append(f.out, '(')
		f.expr(n.Expr)
		f.out = append(f.out, ')')

	case *ast.CallExpr:
		f.expr(n.Func)
		f.out = append(f.out, '(')
		if len(n.Args) > 0 {
			f.expr(n.Args[0])
		}
		for i := 1; i < len(n.Args); i++ {
			f.out = append(f.out, ", "...)
			f.expr(n.Args[i])
		}
		f.out = append(f.out, ')')

	case *ast.PipeExpr:
		f.expr(n.Arg)
		f.out = append(f.out, " | "...)
		// write Ident
		f.out = append(f.out, n.Func.Value...)

	default:
		panic(fmt.Sprintf("format: unexpected expr kind %T", e))
	}
}
