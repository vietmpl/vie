package parser

import (
	"fmt"

	"github.com/vietmpl/vie/ast"
)

// In an [ErrorList], an error is represented by an *Error.
// The position Pos, if valid, points to the beginning of
// the offending token, and the error condition is described
// by Msg.
type Error struct {
	Pos ast.Pos
	Msg string
}

func (e Error) Error() string {
	return e.Msg
}

type ErrorList []*Error

// Add adds an [Error] with given position and error message to an [ErrorList].
func (p *ErrorList) Add(pos ast.Pos, msg string) {
	*p = append(*p, &Error{pos, msg})
}

func (p ErrorList) Error() string {
	switch len(p) {
	case 0:
		return "no errors"
	case 1:
		return p[0].Error()
	}
	return fmt.Sprintf("%s (and %d more errors)", p[0], len(p)-1)
}

func (p ErrorList) Len() int { return len(p) }
