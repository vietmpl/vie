package value

import "fmt"

type Type uint8

const (
	TypeBool Type = iota
	TypeString
)

func (t Type) String() string {
	switch t {
	case TypeBool:
		return "bool"
	case TypeString:
		return "string"
	default:
		panic(fmt.Sprintf("unexpected Type value: %d", t))
	}
}
