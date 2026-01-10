package parse

import (
	"fmt"

	"github.com/vietmpl/vie/token"
)

var binaryOperators = [...]token.Kind{
	token.BANG_EQUAL,
	token.EQUAL_EQUAL,
	token.TILDE,
	token.KEYWORD_AND,
	token.KEYWORD_OR,
}

func parseBinaryOperator(s string) token.Kind {
	for _, kind := range binaryOperators {
		if kind.String() == s {
			return kind
		}
	}
	panic(fmt.Sprintf("unknown binary operator: %s", s))
}

var unaryOperators = [...]token.Kind{
	token.BANG,
}

func parseUnaryOperator(s string) token.Kind {
	for _, kind := range unaryOperators {
		if kind.String() == s {
			return kind
		}
	}
	panic(fmt.Sprintf("unknown unary operator: %s", s))
}
