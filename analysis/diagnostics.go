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
