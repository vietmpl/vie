package analisys

import "fmt"

type Diagnostic interface {
	String() string
	// Pos() token.Pos
}

type WrongUsage struct {
	ExpectedType Type
	GotType      Type
}

func (d *WrongUsage) String() string {
	return fmt.Sprintf("cannot use %s as %s", d.GotType, d.ExpectedType)
}
