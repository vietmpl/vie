package analysis

import (
	"fmt"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/value"
)

type Diagnostic interface {
	String() string
	Pos() ast.Pos
}

type WrongUsage struct {
	WantType value.Type
	GotType  value.Type
	Pos_     ast.Pos
}

func (d *WrongUsage) String() string {
	return fmt.Sprintf("cannot use %s as %s", d.GotType, d.WantType)
}

func (d *WrongUsage) Pos() ast.Pos {
	return d.Pos_
}

type InvalidOperation struct {
	X    value.Type
	Y    value.Type
	Pos_ ast.Pos
}

func (d *InvalidOperation) String() string {
	// TODO(skewb1k): print entire invalid expression (like Go).
	return fmt.Sprintf("invalid operation: mismatched types %s and %s", d.X, d.Y)
}

func (d *InvalidOperation) Pos() ast.Pos {
	return d.Pos_
}

type CrossVarTyping struct {
	X    VarType
	Y    VarType
	Pos_ ast.Pos
}

func (d *CrossVarTyping) String() string {
	return fmt.Sprintf("type of %s depends on type of %s (cross-var typing is not supported yet)", d.X, d.Y)
}

func (d *CrossVarTyping) Pos() ast.Pos {
	return d.Pos_
}

type BuiltinNotFound struct {
	Name string
	Msg  string
	Pos_ ast.Pos
}

func (d *BuiltinNotFound) String() string {
	return d.Msg
}

func (d *BuiltinNotFound) Pos() ast.Pos {
	return d.Pos_
}

type IncorrectArgCount struct {
	FuncName string
	Want     int
	Got      int
	Pos_     ast.Pos
}

func (d *IncorrectArgCount) String() string {
	return fmt.Sprintf("function %q expects %d argument(s), but %d were provided", d.FuncName, d.Want, d.Got)
}

func (d *IncorrectArgCount) Pos() ast.Pos {
	return d.Pos_
}
