// Package token defines constants representing the lexical tokens
// of the Vie template language.
package token

// Kind identifies the type of a lexical token.
type Kind uint8

const (
	ERROR Kind = iota
	EOF
	TEXT
	COMMENT
	IDENTIFIER
	STRING_LITERAL
	L_BRACE_PERCENT
	R_BRACE_PERCENT
	L_BRACE_POUND
	R_BRACE_POUND
	L_DOUBLE_BRACE
	R_DOUBLE_BRACE
	L_PAREN
	R_PAREN
	BANG
	BANG_EQUAL
	COLON
	EQUAL_EQUAL
	PIPE
	QUESTION_MARK
	TILDE
	KEYWORD_AND
	KEYWORD_ELSE
	KEYWORD_ELSEIF
	KEYWORD_END
	KEYWORD_IF
	KEYWORD_OR
)

// String returns the human-readable representation of the token kind.
func (k Kind) String() string {
	return kindToString[k]
}

var kindToString = [...]string{
	ERROR:           "ERROR",
	EOF:             "EOF",
	TEXT:            "TEXT",
	COMMENT:         "COMMENT",
	IDENTIFIER:      "IDENTIFIER",
	STRING_LITERAL:  "STRING_LITERAL",
	L_BRACE_PERCENT: "{%",
	R_BRACE_PERCENT: "%}",
	L_BRACE_POUND:   "{#",
	R_BRACE_POUND:   "#}",
	L_DOUBLE_BRACE:  "{{",
	R_DOUBLE_BRACE:  "}}",
	L_PAREN:         "(",
	R_PAREN:         ")",
	BANG:            "!",
	BANG_EQUAL:      "!=",
	COLON:           ":",
	EQUAL_EQUAL:     "==",
	PIPE:            "|",
	QUESTION_MARK:   "?",
	TILDE:           "~",
	KEYWORD_AND:     "and",
	KEYWORD_ELSE:    "else",
	KEYWORD_ELSEIF:  "elseif",
	KEYWORD_END:     "end",
	KEYWORD_IF:      "if",
	KEYWORD_OR:      "or",
}
