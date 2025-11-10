package value

import "fmt"

type Type uint8

const (
	TypeString Type = iota
	TypeBool
	TypeFunction
)

func (t Type) String() string {
	switch t {
	case TypeString:
		return "string"
	case TypeBool:
		return "bool"
	case TypeFunction:
		return "function"
	default:
		panic(fmt.Sprintf("unexpected Type value: %d", t))
	}
}
