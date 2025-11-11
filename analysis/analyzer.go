package analysis

import (
	"fmt"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/builtin"
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
		a.stmt(s)
	}
}

func (a *analyzer) stmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.Text:
		// Skip
	case *ast.RenderStmt:
		x := a.expr(s.X)
		switch xx := x.(type) {
		case value.Type:
			if xx != value.TypeString {
				a.diagnostics = append(a.diagnostics, &WrongUsage{
					WantType: value.TypeString,
					GotType:  xx,
					Pos_:     s.X.Pos(),
				})
			}
		case VarType:
			a.addUsage(xx.String(), Usage{
				Type: value.TypeString,
				Kind: UsageKindRender,
				Pos:  s.X.Pos(),
			})
		}

	case *ast.IfStmt:
		cond := a.expr(s.Cond)
		switch condx := cond.(type) {
		case value.Type:
			if condx != value.TypeBool {
				a.diagnostics = append(a.diagnostics, &WrongUsage{
					WantType: value.TypeBool,
					GotType:  condx,
					Pos_:     s.Cond.Pos(),
				})
			}
		case VarType:
			a.addUsage(condx.String(), Usage{
				Type: value.TypeBool,
				Kind: UsageKindIf,
				Pos:  s.Cond.Pos(),
			})
		}
		a.stmts(s.Cons)
		for _, elseIfClause := range s.ElseIfs {
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
		if s.Else != nil {
			a.stmts(s.Else.Cons)
		}

	// case *ast.SwitchStmt:

	default:
		panic(fmt.Sprintf("analyzer: unexpected stmt type %T", stmt))
	}
}

// TODO(skewb1k): avoid using any to represent `[value.Type] | [VarType]`.
func (a *analyzer) expr(expr ast.Expr) any {
	switch e := expr.(type) {
	case *ast.BasicLit:
		switch e.Kind {
		case ast.KindBool:
			return value.TypeBool
		case ast.KindString:
			return value.TypeString
		default:
			panic(fmt.Sprintf("analyzer: unexpected BasicLit kind %d", e.Kind))
		}

	case *ast.Ident:
		return VarType(e.Name)

	case *ast.UnaryExpr:
		x := a.expr(e.X)
		// The '!' and 'not' operators can only be applied to boolean values
		a.expectType(x, Usage{
			Type: value.TypeBool,
			Kind: UsageKindUnOp,
			Pos:  e.X.Pos(),
		})
		return value.TypeBool

	case *ast.BinaryExpr:
		switch e.Op {
		case ast.BinOpKindConcat:
			x := a.expr(e.X)
			a.expectType(x, Usage{
				Type: value.TypeString,
				Kind: UsageKindBinOp,
				Pos:  e.X.Pos(),
			})

			y := a.expr(e.Y)
			a.expectType(y, Usage{
				Type: value.TypeString,
				Kind: UsageKindBinOp,
				Pos:  e.Y.Pos(),
			})
			return value.TypeString

		case
			ast.BinOpKindEq,
			ast.BinOpKindNeq,
			ast.BinOpKindIs,
			ast.BinOpKindIsNot:

			x := a.expr(e.X)
			y := a.expr(e.Y)
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
							Pos_: e.Pos(),
						})
					}
				// <lit> is <var>
				case VarType:
					a.addUsage(yy.String(), Usage{
						Type: xx,
						Kind: UsageKindBinOp,
						Pos:  e.Pos(),
					})
				}
			case VarType:
				switch yy := y.(type) {
				// <var> is <lit>
				case value.Type:
					a.addUsage(xx.String(), Usage{
						Type: yy,
						Kind: UsageKindBinOp,
						Pos:  e.Pos(),
					})
				// <var> is <var>
				case VarType:
					a.diagnostics = append(a.diagnostics, &CrossVarTyping{
						X:    xx,
						Y:    yy,
						Pos_: e.Pos(),
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

			x := a.expr(e.X)
			a.expectType(x, Usage{
				Type: value.TypeBool,
				Kind: UsageKindBinOp,
				Pos:  e.X.Pos(),
			})

			y := a.expr(e.Y)
			a.expectType(y, Usage{
				Type: value.TypeBool,
				Kind: UsageKindBinOp,
				Pos:  e.Y.Pos(),
			})
			return value.TypeBool

		default:
			panic(fmt.Sprintf("analyzer: unexpected BinOpKind: %T", expr))
		}

	case *ast.ParenExpr:
		return a.expr(e.X)

	case *ast.CallExpr:
		fn, err := builtin.LookupFunction(e.Func)
		if err != nil {
			a.diagnostics = append(a.diagnostics, &BuiltinNotFound{
				Name: e.Func.Name,
				Msg:  err.Error(),
				Pos_: e.Pos(),
			})
			// TODO(skewb1k): return nil.
			return value.TypeString
		}
		if len(e.Args) != len(fn.ArgTypes) {
			a.diagnostics = append(a.diagnostics, &IncorrectArgCount{
				FuncName: e.Func.Name,
				Got:      len(e.Args),
				Want:     len(fn.ArgTypes),
				Pos_:     e.Pos(),
			})
			return fn.ReturnType
		}
		for i, arg := range e.Args {
			x := a.expr(arg)
			a.expectType(x, Usage{
				Type: fn.ArgTypes[i],
				Kind: UsageKindCall,
				Pos:  arg.Pos(),
			})
		}
		return fn.ReturnType
	// case *ast.PipeExpr:

	default:
		panic(fmt.Sprintf("analyzer: unexpected expr type %T", expr))
	}
}

func (a *analyzer) expectType(x any, u Usage) {
	switch xx := x.(type) {
	case value.Type:
		if xx != u.Type {
			a.diagnostics = append(a.diagnostics, &WrongUsage{
				WantType: u.Type,
				GotType:  xx,
				Pos_:     u.Pos,
			})
		}
	case VarType:
		a.addUsage(xx.String(), u)
	}
}
