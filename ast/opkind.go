package ast

import "fmt"

type BinOpKind uint8

const (
	BinOpKindEq BinOpKind = iota
	BinOpKindNeq
	BinOpKindLss
	BinOpKindLeq
	BinOpKindGtr
	BinOpKindGeq
	BinOpKindIs
	BinOpKindIsNot
	BinOpKindOr
	BinOpKindAnd
	BinOpKindLOr
	BinOpKindLAnd
	BinOpKindConcat
)

var binOpStrings = [...]string{
	"==",
	"!=",
	"<",
	"<=",
	">",
	">=",
	"is",
	"is not",
	"or",
	"and",
	"||",
	"&&",
	"~",
}

func (k BinOpKind) String() string {
	return binOpStrings[k]
}
func ParseBinOpKind(s string) BinOpKind {
	for i, str := range binOpStrings {
		if str == s {
			return BinOpKind(i)
		}
	}
	panic(fmt.Sprintf("unexpected UnOpKind string: %s", s))
}

type UnOpKind uint8

const (
	UnOpKindExcl UnOpKind = iota
	UnOpKindNot
)

var unOpStrings = [...]string{
	"!",
	"not",
}

func (k UnOpKind) String() string {
	return unOpStrings[k]
}
func ParseUnOpKind(s string) UnOpKind {
	for i, str := range unOpStrings {
		if str == s {
			return UnOpKind(i)
		}
	}
	panic(fmt.Sprintf("unexpected UnOpKind string: %s", s))
}
