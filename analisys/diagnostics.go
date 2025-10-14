package analisys

import (
	"fmt"

	"github.com/vietmpl/vie/ast"
)

type Diagnostic interface {
	String() string
	Pos() ast.Pos
}

type WrongUsage struct {
	ExpectedType Type
	GotType      Type
	_Pos         ast.Pos
}

func (d *WrongUsage) String() string {
	return fmt.Sprintf("cannot use %s as %s", d.GotType, d.ExpectedType)
}

func (d *WrongUsage) Pos() ast.Pos {
	return d._Pos
}

type InvalidOperation struct {
	Typ1 Type
	Typ2 Type
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
