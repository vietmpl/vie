package analysis

import (
	"fmt"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/value"
)

type VarType string

func (vt VarType) String() string {
	return string(vt)
}

type analyzer struct {
	usages      map[string][]Usage
	diagnostics []Diagnostic
}

func (a *analyzer) addUsage(varName string, u Usage) {
	a.usages[varName] = append(a.usages[varName], u)
}

func File(file *ast.File) (map[string]value.Type, []Diagnostic) {
	a := analyzer{
		usages: make(map[string][]Usage),
	}
	a.stmts(file.Stmts)

	types := make(map[string]value.Type, len(a.usages))
	for name, uses := range a.usages {
		var typeCount [value.TypeFunction]uint
		for _, u := range uses {
			typeCount[u.Type]++
		}

		var maxType value.Type
		var maxCount uint
		for t, c := range typeCount {
			if c > maxCount {
				maxType, maxCount = value.Type(t), c
			}
		}

		types[name] = maxType

		// report mismatched usages
		for _, u := range uses {
			if u.Type != maxType {
				a.diagnostics = append(a.diagnostics, &WrongUsage{
					WantType: u.Type,
					GotType:  maxType,
					Pos_:     u.Pos,
				})
			}
		}
	}
	return types, a.diagnostics
}

func (a *analyzer) stmts(stmts []ast.Stmt) {
	for _, s := range stmts {
		switch n := s.(type) {
		case *ast.Text:
			// Skip
		case *ast.RenderStmt:
			x := a.expr(n.X)
			switch xx := x.(type) {
			case value.Type:
				if xx != value.TypeString {
					a.diagnostics = append(a.diagnostics, &WrongUsage{
						WantType: value.TypeString,
						GotType:  xx,
						Pos_:     n.X.Pos(),
					})
				}
			case VarType:
				a.addUsage(xx.String(), Usage{
					Type: value.TypeString,
					Kind: UsageKindRender,
					Pos:  n.X.Pos(),
				})
			}

		case *ast.IfStmt:
			cond := a.expr(n.Cond)
			switch condx := cond.(type) {
			case value.Type:
				if condx != value.TypeBool {
					a.diagnostics = append(a.diagnostics, &WrongUsage{
						WantType: value.TypeBool,
						GotType:  condx,
						Pos_:     n.Cond.Pos(),
					})
				}
			case VarType:
				a.addUsage(condx.String(), Usage{
					Type: value.TypeBool,
					Kind: UsageKindIf,
					Pos:  n.Cond.Pos(),
				})
			}
			a.stmts(n.Cons)
			for _, elseIfClause := range n.ElseIfs {
				elseIfCond := a.expr(elseIfClause.Cond)
				switch elseIfCondx := elseIfCond.(type) {
				case value.Type:
					if elseIfCondx != value.TypeBool {
						a.diagnostics = append(a.diagnostics, &WrongUsage{
							WantType: value.TypeBool,
							GotType:  elseIfCondx,
							Pos_:     elseIfClause.Cond.Pos(),
						})
					}
				case VarType:
					a.addUsage(elseIfCondx.String(), Usage{
						Type: value.TypeBool,
						Kind: UsageKindIf,
						Pos:  elseIfClause.Cond.Pos(),
					})
				}
				a.stmts(elseIfClause.Cons)
			}
			if n.Else != nil {
				a.stmts(n.Else.Cons)
			}

		// case *ast.SwitchStmt:

		default:
			panic(fmt.Sprintf("analyzer: unexpected stmt type %T", s))
		}
	}
}

func (a *analyzer) expr(e ast.Expr) any {
	switch n := e.(type) {
	case *ast.BasicLit:
		switch n.Kind {
		case ast.KindBool:
			return value.TypeBool
		case ast.KindString:
			return value.TypeString
		default:
			panic(fmt.Sprintf("analyzer: unexpected BasicLit kind %d", n.Kind))
		}

	case *ast.Ident:
		return VarType(n.Name)

	case *ast.UnaryExpr:
		x := a.expr(n.X)
		// The '!' and 'not' operators can only be applied to boolean values
		return x

	case *ast.BinaryExpr:
		switch n.Op {
		case ast.BinOpKindConcat:
			x := a.expr(n.X)
			a.expectOperandType(x, n.X.Pos(), value.TypeString)

			y := a.expr(n.Y)
			a.expectOperandType(y, n.Y.Pos(), value.TypeString)
			return value.TypeString

		case
			ast.BinOpKindEq,
			ast.BinOpKindNeq,
			ast.BinOpKindIs,
			ast.BinOpKindIsNot:

			x := a.expr(n.X)
			y := a.expr(n.Y)
			// TODO(skewb1k): refactor.
			switch xx := x.(type) {
			case value.Type:
				switch yy := y.(type) {
				// <lit> is <lit>
				case value.Type:
					// catch `false is "str"`
					if xx != yy {
						a.diagnostics = append(a.diagnostics, &InvalidOperation{
							X:    xx,
							Y:    yy,
							Pos_: n.Pos(),
						})
					}
				// <lit> is <var>
				case VarType:
					a.addUsage(yy.String(), Usage{
						Type: xx,
						Kind: UsageKindBinop,
						Pos:  n.Pos(),
					})
				}
			case VarType:
				switch yy := y.(type) {
				// <var> is <lit>
				case value.Type:
					a.addUsage(xx.String(), Usage{
						Type: yy,
						Kind: UsageKindBinop,
						Pos:  n.Pos(),
					})
				// <var> is <var>
				case VarType:
					a.diagnostics = append(a.diagnostics, &CrossVarTyping{
						X:    xx,
						Y:    yy,
						Pos_: n.Pos(),
					})
				}
			}
			return value.TypeBool

		case ast.BinOpKindGtr,
			ast.BinOpKindGeq,
			ast.BinOpKindLss,
			ast.BinOpKindLeq,
			ast.BinOpKindLAnd,
			ast.BinOpKindLOr,
			ast.BinOpKindAnd,
			ast.BinOpKindOr:

			x := a.expr(n.X)
			a.expectOperandType(x, n.X.Pos(), value.TypeBool)

			y := a.expr(n.Y)
			a.expectOperandType(y, n.Y.Pos(), value.TypeBool)
			return value.TypeBool

		default:
			panic(fmt.Sprintf("analyzer: unexpected BinOpKind: %T", e))
		}

	case *ast.ParenExpr:
		return a.expr(n.X)

	// case *ast.CallExpr:
	// case *ast.PipeExpr:

	default:
		panic(fmt.Sprintf("analyzer: unexpected expr type %T", e))
	}
}

func (a *analyzer) expectOperandType(x any, pos ast.Pos, typ value.Type) {
	switch xx := x.(type) {
	case value.Type:
		if xx != typ {
			a.diagnostics = append(a.diagnostics, &WrongUsage{
				WantType: typ,
				GotType:  xx,
				Pos_:     pos,
			})
		}
	case VarType:
		a.addUsage(xx.String(), Usage{
			Type: typ,
			Kind: UsageKindBinop,
			Pos:  pos,
		})
	}
}
