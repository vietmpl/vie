package analysis

import (
	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/value"
)

type usageKind uint8

const (
	UsageKindRender usageKind = iota
	UsageKindSwitch
	UsageKindIf
	UsageKindBinOp
	UsageKindUnOp
)

type Usage struct {
	Type value.Type
	Kind usageKind
	Pos  ast.Pos
}
