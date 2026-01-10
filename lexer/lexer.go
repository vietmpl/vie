package lexer

import (
	"github.com/vietmpl/vie/token"
)

type Token struct {
	Kind  token.Kind
	Start int
	End   int
}

var keywords = map[string]token.Kind{
	"and":    token.KEYWORD_AND,
	"else":   token.KEYWORD_ELSE,
	"elseif": token.KEYWORD_ELSEIF,
	"end":    token.KEYWORD_END,
	"if":     token.KEYWORD_IF,
	"or":     token.KEYWORD_OR,
}

type state int

const (
	start state = iota
	startExpression
	startComment
	text
	comment
	identifier
	stringLiteral
	lBrace
	rBrace
	bang
	equal
	percent
	pound
)

type Lexer struct {
	source []byte
	index  int
	state  state
}

func (l *Lexer) Init(source []byte) {
	l.source = append(source, 0)
	// TODO(skewb1k): handle the UTF-8 BOM.
	l.index = 0
	l.state = start
}

func (l *Lexer) Next() Token {
	result := Token{
		Kind:  token.EOF,
		Start: l.index,
		End:   0,
	}

loop:
	for {
		switch l.state {
		case start:
			switch l.source[l.index] {
			case 0:
				// TODO(skewb1k): handle if we are not at the end of the buffer.
				break loop
			case '{':
				switch l.source[l.index+1] {
				case '{', '%', '#':
					l.state = lBrace
				default:
					result.Kind = token.TEXT
					l.state = text
				}
			default:
				result.Kind = token.TEXT
				l.state = text
			}

		case text:
			l.index++
			switch l.source[l.index] {
			case 0, '\n':
				l.state = start
				break loop
			case '{':
				switch l.source[l.index+1] {
				case '{', '%', '#':
					l.state = lBrace
					break loop
				}
			}

		case lBrace:
			l.index++
			switch l.source[l.index] {
			case '{':
				result.Kind = token.L_DOUBLE_BRACE
				l.state = startExpression
			case '%':
				result.Kind = token.L_BRACE_PERCENT
				l.state = startExpression
			case '#':
				result.Kind = token.L_BRACE_POUND
				l.state = startComment
			default:
				panic("unreachable")
			}
			l.index++
			break loop

		case rBrace:
			l.index++
			switch l.source[l.index] {
			case '}':
				l.index++
				result.Kind = token.R_DOUBLE_BRACE
			default:
				result.Kind = token.ERROR
			}
			l.state = start
			break loop

		case percent:
			l.index++
			switch l.source[l.index] {
			case '}':
				l.index++
				result.Kind = token.R_BRACE_PERCENT
			default:
				result.Kind = token.ERROR
			}
			l.state = start
			break loop

		case pound:
			l.index++
			switch l.source[l.index] {
			case '}':
				l.index++
				result.Kind = token.R_BRACE_POUND
			default:
				result.Kind = token.ERROR
			}
			l.state = start
			break loop

		case startComment:
			switch l.source[l.index] {
			case 0, '\n':
				result.Kind = token.ERROR
				l.index++
				l.state = startComment
				break loop

			case '#':
				switch l.source[l.index+1] {
				case '}':
					l.state = pound
				}
			default:
				result.Kind = token.COMMENT
				l.state = comment
			}

		case comment:
			l.index++
			switch l.source[l.index] {
			case 0:
				l.state = start
				break loop
			case '\n':
				l.state = startComment
				break loop
			case '#':
				switch l.source[l.index+1] {
				case '}':
					l.state = pound
					break loop
				}
			}

		case startExpression:
			ch := l.source[l.index]
			switch ch {
			case 0:
				l.state = start
				break loop
			case '\n':
				result.Kind = token.ERROR
				l.index++
				l.state = startExpression
				break loop
			case '}':
				l.state = rBrace
			case '%':
				l.state = percent
			case '#':
				l.state = percent
			case '"':
				result.Kind = token.STRING_LITERAL
				l.state = stringLiteral
			case '!':
				l.state = bang
			case '=':
				l.state = equal
			// case ':':
			// 	l.index++
			// 	result.Kind = COLON
			// 	break loop
			// case '?':
			// 	l.index++
			// 	result.Kind = QUESTION_MARK
			// 	break loop
			case '|':
				l.index++
				result.Kind = token.PIPE
				break loop
			case '~':
				l.index++
				result.Kind = token.TILDE
				break loop
			case '(':
				l.index++
				result.Kind = token.L_PAREN
				break loop
			case ')':
				l.index++
				result.Kind = token.R_PAREN
				break loop
			default:
				switch {
				case isSpace(ch):
					l.index++
					result.Start = l.index
				case isLetter(ch) || ch == '_':
					result.Kind = token.IDENTIFIER
					l.state = identifier
				default:
					l.index++
					result.Kind = token.ERROR
					break loop
				}
			}

		case identifier:
			l.index++
			switch ch := l.source[l.index]; {
			case isLetter(ch) || isDigit(ch) || ch == '_':
			default:
				id := l.source[result.Start:l.index]
				if kind, ok := keywords[string(id)]; ok {
					result.Kind = kind
				}
				l.state = startExpression
				break loop
			}

		case stringLiteral:
			l.index++
			switch l.source[l.index] {
			case 0, '\n':
				result.Kind = token.ERROR
				l.state = start
				break loop
			case '\\':
				panic("TODO: support escaping")
			case '"':
				l.index++
				l.state = startExpression
				break loop
			}

		case bang:
			l.index++
			switch l.source[l.index] {
			case '=':
				l.index++
				result.Kind = token.BANG_EQUAL
			default:
				result.Kind = token.BANG
			}
			l.state = startExpression
			break loop

		case equal:
			l.index++
			switch l.source[l.index] {
			case '=':
				l.index++
				result.Kind = token.EQUAL_EQUAL
			default:
				result.Kind = token.ERROR
			}
			l.state = startExpression
			break loop
		}
	}
	result.End = l.index
	return result
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\t' || b == '\r'
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
