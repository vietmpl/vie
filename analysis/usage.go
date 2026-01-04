package analysis

import (
	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/value"
)

type usageKind uint8

const (
	UsageKindRender usageKind = iota
	UsageKindIf
	UsageKindBinOp
	UsageKindUnOp
	UsageKindCall
)

type Usage struct {
	Type value.Type
	Kind usageKind
	Pos  ast.Location
	Path string
}
