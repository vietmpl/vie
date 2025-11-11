package analysis

import (
	"fmt"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/builtin"
	"github.com/vietmpl/vie/value"
)

type analyzer struct {
	usages      map[string][]Usage
	diagnostics []Diagnostic
}

func (a *analyzer) addUsage(varName string, u Usage) {
	a.usages[varName] = append(a.usages[varName], u)
}

func (a *analyzer) addDiagnostic(d Diagnostic) {
	a.diagnostics = append(a.diagnostics, d)
}

func CheckFile(file *ast.File) (map[string]value.Type, []Diagnostic) {
	a := analyzer{
		usages: make(map[string][]Usage),
	}
	a.checkStmts(file.Stmts)

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
				a.addDiagnostic(WrongUsage{
					WantType: u.Type,
					GotType:  maxType,
					Pos_:     u.Pos,
				})
			}
		}
	}
	return types, a.diagnostics
}

func (a *analyzer) checkStmts(stmts []ast.Stmt) {
	for _, s := range stmts {
		a.checkStmt(s)
	}
}

func (a *analyzer) checkStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.Text:
		// Skip
	case *ast.RenderStmt:
		x := a.checkExpr(s.X)
		switch xx := x.(type) {
		case nil:
			return
		case value.Type:
			if xx != value.TypeString {
				a.addDiagnostic(WrongUsage{
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
		cond := a.checkExpr(s.Cond)
		switch condx := cond.(type) {
		case nil:
			return
		case value.Type:
			if condx != value.TypeBool {
				a.addDiagnostic(WrongUsage{
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
		a.checkStmts(s.Cons)
		for _, elseIfClause := range s.ElseIfs {
			elseIfCond := a.checkExpr(elseIfClause.Cond)
			switch elseIfCondx := elseIfCond.(type) {
			case nil:
				return
			case value.Type:
				if elseIfCondx != value.TypeBool {
					a.addDiagnostic(WrongUsage{
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
			a.checkStmts(elseIfClause.Cons)
		}
		if s.Else != nil {
			a.checkStmts(s.Else.Cons)
		}

	// case *ast.SwitchStmt:

	default:
		panic(fmt.Sprintf("analyzer: unexpected stmt type %T", stmt))
	}
}

// TODO(skewb1k): avoid using any to represent `[value.Type] | [VarType]`.
func (a *analyzer) checkExpr(expr ast.Expr) any {
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
		x := a.checkExpr(e.X)
		if x == nil {
			return nil
		}
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
			x := a.checkExpr(e.X)
			y := a.checkExpr(e.Y)
			if x == nil || y == nil {
				return nil
			}

			a.expectType(x, Usage{
				Type: value.TypeString,
				Kind: UsageKindBinOp,
				Pos:  e.X.Pos(),
			})
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

			x := a.checkExpr(e.X)
			y := a.checkExpr(e.Y)
			if x == nil || y == nil {
				return nil
			}
			// TODO(skewb1k): refactor.
			switch xx := x.(type) {
			case value.Type:
				switch yy := y.(type) {
				// <lit> is <lit>
				case value.Type:
					// catch `false is "str"`
					if xx != yy {
						a.addDiagnostic(InvalidOperation{
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
					a.addDiagnostic(CrossVarTyping{
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

			x := a.checkExpr(e.X)
			y := a.checkExpr(e.Y)
			if x == nil || y == nil {
				return nil
			}

			a.expectType(x, Usage{
				Type: value.TypeBool,
				Kind: UsageKindBinOp,
				Pos:  e.X.Pos(),
			})
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
		return a.checkExpr(e.X)

	case *ast.CallExpr:
		return a.checkFunc(e.Func, e.Args)

	case *ast.PipeExpr:
		return a.checkFunc(e.Func, []ast.Expr{e.Arg})

	default:
		panic(fmt.Sprintf("analyzer: unexpected expr type %T", expr))
	}
}

func (a *analyzer) checkFunc(ident ast.Ident, exprs []ast.Expr) any {
	fn, err := builtin.LookupFunction(ident)
	if err != nil {
		a.addDiagnostic(BuiltinNotFound{
			Name: ident.Name,
			Msg:  err.Error(),
			Pos_: ident.Pos(),
		})
		return nil
	}
	// TODO(skewb1k): improve error messages for PipeExpr.
	if len(exprs) != len(fn.ArgTypes) {
		a.addDiagnostic(IncorrectArgCount{
			FuncName: ident.Name,
			Got:      len(exprs),
			Want:     len(fn.ArgTypes),
			// TODO(skewb1k): use proper arg pos.
			Pos_: ident.Pos(),
		})
		return fn.ReturnType
	}

	// Evaluate and collect argument types for the function call. If any
	// argument expression cannot be typed (returns nil), stop processing
	// and propagate nil.
	type typedArg struct {
		typ  any
		expr ast.Expr
	}
	args := make([]typedArg, 0, len(exprs))
	for _, arg := range exprs {
		x := a.checkExpr(arg)
		if x == nil {
			return nil
		}
		args = append(args, typedArg{
			typ:  x,
			expr: arg,
		})
	}

	for i, arg := range args {
		a.expectType(arg.typ, Usage{
			Type: fn.ArgTypes[i],
			Kind: UsageKindCall,
			Pos:  arg.expr.Pos(),
		})
	}
	return fn.ReturnType
}

func (a *analyzer) expectType(x any, u Usage) {
	switch xx := x.(type) {
	case value.Type:
		if xx != u.Type {
			a.addDiagnostic(WrongUsage{
				WantType: u.Type,
				GotType:  xx,
				Pos_:     u.Pos,
			})
		}
	case VarType:
		a.addUsage(xx.String(), u)
	}
}

type VarType string

func (vt VarType) String() string {
	return string(vt)
}
