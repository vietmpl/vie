package analisys

import (
	"fmt"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/value"
)

type analyzer struct {
	tm          map[string]*[value.TypeString + 1]uint
	diagnostics []Diagnostic
}

func newAnalyzer() *analyzer {
	return &analyzer{
		tm: make(map[string]*[value.TypeString + 1]uint),
	}
}

func (a *analyzer) addType(name string, typ value.Type) {
	if a.tm[name] == nil {
		a.tm[name] = new([value.TypeString + 1]uint)
	}

	a.tm[name][typ]++
}

func File(file *ast.File) (map[string]value.Type, []Diagnostic) {
	a := newAnalyzer()
	a.stmts(file.Stmts)

	types := make(map[string]value.Type, len(a.tm))
	for varname, usages := range a.tm {
		var maxCount uint
		var maxType value.Type
		for t, count := range usages {
			if count > maxCount {
				maxCount = count
				maxType = value.Type(t)
			}
		}
		types[varname] = maxType
		// TODO: report wrong usages diagnostics
	}
	return types, a.diagnostics
}

func (a *analyzer) stmts(stmts []ast.Stmt) {
	for _, s := range stmts {
		switch n := s.(type) {
		case *ast.Text:

		case *ast.RenderStmt:
			a.expr(n.X, value.TypeString)

		case *ast.IfStmt:
			a.expr(n.Cond, value.TypeBool)
			a.stmts(n.Cons)
			for _, elseIfClause := range n.ElseIfs {
				a.expr(elseIfClause.Cond, value.TypeBool)
				a.stmts(elseIfClause.Cons)
			}
			if n.Else != nil {
				a.stmts(n.Else.Cons)
			}

		case *ast.SwitchStmt:

		default:
			panic(fmt.Sprintf("analyzer: unexpected stmt type %T", s))
		}
	}
}

func (a *analyzer) expr(e ast.Expr, typ value.Type) {
	switch n := e.(type) {
	case *ast.BasicLit:
		var got value.Type
		switch n.Kind {
		case ast.KindBool:
			got = value.TypeBool
		case ast.KindString:
			got = value.TypeString
		default:
			panic(fmt.Sprintf("analyzer: unexpected BasicLit kind %d", n.Kind))
		}

		if got != typ {
			a.diagnostics = append(a.diagnostics, &WrongUsage{
				ExpectedType: typ,
				GotType:      got,
				_Pos:         n.Pos(),
			})
		}

	case *ast.Ident:
		a.addType(string(n.Name), typ)

	case *ast.UnaryExpr:
		// The '!' and 'not' operators can only be applied to boolean values
		a.expr(n.X, value.TypeBool)

	case *ast.BinaryExpr:
		switch n.Op {
		case ast.BinOpKindConcat:
			a.expr(n.X, value.TypeString)
			a.expr(n.Y, value.TypeString)
		case
			ast.BinOpKindEq,
			ast.BinOpKindNeq,
			ast.BinOpKindIs,
			ast.BinOpKindIsNot,

			ast.BinOpKindGtr,
			ast.BinOpKindGeq,
			ast.BinOpKindLss,
			ast.BinOpKindLeq,
			ast.BinOpKindLAnd,
			ast.BinOpKindLOr,
			ast.BinOpKindAnd,
			ast.BinOpKindOr:
			a.expr(n.X, value.TypeBool)
			a.expr(n.Y, value.TypeBool)
		}

	case *ast.ParenExpr:
		a.expr(n.X, typ)

	case *ast.CallExpr:
	case *ast.PipeExpr:

	default:
		panic(fmt.Sprintf("analyzer: unexpected expr type %T", e))
	}
}
