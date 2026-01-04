package ast

import "fmt"

type BinaryOperator uint8

const (
	EQUAL BinaryOperator = iota
	NOT_EQUAL
	OR
	AND
	CONCAT
)

var binaryOperatorStrings = [...]string{
	EQUAL:     "==",
	NOT_EQUAL: "!=",
	OR:        "or",
	AND:       "and",
	CONCAT:    "~",
}

func (k BinaryOperator) String() string {
	return binaryOperatorStrings[k]
}

func ParseBinaryOperator(s string) BinaryOperator {
	for i, str := range binaryOperatorStrings {
		if str == s {
			return BinaryOperator(i)
		}
	}
	panic(fmt.Sprintf("unexpected binary operator string: %s", s))
}

type UnaryOperator uint8

const (
	NOT UnaryOperator = iota
)

var unaryOperatorStrings = [...]string{
	NOT: "!",
}

func (k UnaryOperator) String() string {
	return unaryOperatorStrings[k]
}

func ParseUnaryOperator(s string) UnaryOperator {
	for i, str := range unaryOperatorStrings {
		if str == s {
			return UnaryOperator(i)
		}
	}
	panic(fmt.Sprintf("unexpected unary operator string: %s", s))
}
