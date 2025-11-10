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
	ExpectedType value.Type
	GotType      value.Type
	_Pos         ast.Pos
}

func (d *WrongUsage) String() string {
	return fmt.Sprintf("cannot use %s as %s", d.GotType, d.ExpectedType)
}

func (d *WrongUsage) Pos() ast.Pos {
	return d._Pos
}

type InvalidOperation struct {
	Typ1 value.Type
	Typ2 value.Type
	_Pos ast.Pos
}

func (d *InvalidOperation) String() string {
	// TODO: print entire invalid expression (like Go).
	return fmt.Sprintf("invalid operation: mismatched types %s and %s", d.Typ1, d.Typ2)
}

func (d *InvalidOperation) Pos() ast.Pos {
	return d._Pos
}

type NonBoolealInIf struct {
	_Pos ast.Pos
}

func (d *NonBoolealInIf) String() string {
	return "non-boolean condition in if statement"
}

func (d *NonBoolealInIf) Pos() ast.Pos {
	return d._Pos
}
