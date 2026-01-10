package lexer_test

import (
	"testing"

	"github.com/vietmpl/vie/lexer"
	. "github.com/vietmpl/vie/token"
)

func TestLexerPositions(t *testing.T) {
	cases := []struct {
		source string
		tokens []lexer.Token
	}{
		{"a", []lexer.Token{
			{TEXT, 0, 1},
		}},
		{"{{id", []lexer.Token{
			{L_DOUBLE_BRACE, 0, 2},
			{IDENTIFIER, 2, 4},
		}},
		// TODO(skewb1k): handle Unicode.
		{"{{idλ", []lexer.Token{
			{L_DOUBLE_BRACE, 0, 2},
			{IDENTIFIER, 2, 4},
			{ERROR, 4, 5},
			{ERROR, 5, 6},
		}},
		{"{{\n}}", []lexer.Token{
			{L_DOUBLE_BRACE, 0, 2},
			{ERROR, 2, 3},
			{R_DOUBLE_BRACE, 3, 5},
		}},
		{"{# comment #}", []lexer.Token{
			{L_BRACE_POUND, 0, 2},
			{COMMENT, 2, 11},
			{R_BRACE_POUND, 11, 13},
		}},
		{"{#start\n#}", []lexer.Token{
			{L_BRACE_POUND, 0, 2},
			{COMMENT, 2, 7},
			{ERROR, 7, 8},
			{R_BRACE_POUND, 8, 10},
		}},
		{"aa{{ id }}bb", []lexer.Token{
			{TEXT, 0, 2},
			{L_DOUBLE_BRACE, 2, 4},
			{IDENTIFIER, 5, 7},
			{R_DOUBLE_BRACE, 8, 10},
			{TEXT, 10, 12},
		}},
		{"aa{% if id %}bb", []lexer.Token{
			{TEXT, 0, 2},
			{L_BRACE_PERCENT, 2, 4},
			{KEYWORD_IF, 5, 7},
			{IDENTIFIER, 8, 10},
			{R_BRACE_PERCENT, 11, 13},
			{TEXT, 13, 15},
		}},
		{"{% if} text", []lexer.Token{
			{L_BRACE_PERCENT, 0, 2},
			{KEYWORD_IF, 3, 5},
			{ERROR, 5, 6},
			{TEXT, 6, 11},
		}},
		{`{{ "hi" "1""2"}}`, []lexer.Token{
			{L_DOUBLE_BRACE, 0, 2},
			{STRING_LITERAL, 3, 7},
			{STRING_LITERAL, 8, 11},
			{STRING_LITERAL, 11, 14},
			{R_DOUBLE_BRACE, 14, 16},
		}},
		{`{{ or }}`, []lexer.Token{
			{L_DOUBLE_BRACE, 0, 2},
			{KEYWORD_OR, 3, 5},
			{R_DOUBLE_BRACE, 6, 8},
		}},
		{`{{ (( ) }}`, []lexer.Token{
			{L_DOUBLE_BRACE, 0, 2},
			{L_PAREN, 3, 4},
			{L_PAREN, 4, 5},
			{R_PAREN, 6, 7},
			{R_DOUBLE_BRACE, 8, 10},
		}},
	}
	for _, tt := range cases {
		t.Run(tt.source, func(t *testing.T) {
			var l lexer.Lexer
			l.Init([]byte(tt.source))
			for i, expected := range tt.tokens {
				actual := l.Next()
				if expected.Kind != actual.Kind {
					t.Errorf("token %d kind mismatch: expected %s, got %s",
						i+1, expected.Kind, actual.Kind)
				}

				if expected.Start != actual.Start {
					t.Errorf("token %d start mismatch: expected %d, got %d",
						i+1, expected.Start, actual.Start)
				}

				if expected.End != actual.End {
					t.Errorf("token %d end mismatch: expected %d, got %d",
						i+1, expected.End, actual.End)
				}
			}
			expectEOF(t, &l, len(tt.source))
		})
	}
}

func TestLexerOperators(t *testing.T) {
	testTokens(t, `{{ ! != == | ~ }}`, []Kind{
		L_DOUBLE_BRACE,
		BANG,
		BANG_EQUAL,
		EQUAL_EQUAL,
		PIPE,
		TILDE,
		R_DOUBLE_BRACE,
	})
}

func TestLexerKeywords(t *testing.T) {
	testTokens(t, `{{ and else elseif end if or }}`, []Kind{
		L_DOUBLE_BRACE,
		KEYWORD_AND,
		KEYWORD_ELSE,
		KEYWORD_ELSEIF,
		KEYWORD_END,
		KEYWORD_IF,
		KEYWORD_OR,
		R_DOUBLE_BRACE,
	})
}

func testTokens(t *testing.T, source string, expectedTokenKinds []Kind) {
	t.Helper()

	var l lexer.Lexer
	l.Init([]byte(source))
	for i, expectedKind := range expectedTokenKinds {
		actual := l.Next()
		if expectedKind != actual.Kind {
			t.Errorf("token %d kind mismatch: expected %s, got %s",
				i+1, expectedKind, actual.Kind)
		}
	}
	expectEOF(t, &l, len(source))
}

func expectEOF(t *testing.T, l *lexer.Lexer, sourceLen int) {
	t.Helper()

	lastToken := l.Next()
	if lastToken.Kind != EOF {
		t.Errorf("last token kind mismatch: expected EOF, got %v",
			lastToken.Kind)
	}
	if lastToken.Start != sourceLen {
		t.Errorf("last token start mismatch: expected %d, got %d",
			sourceLen, lastToken.Start)
	}
	if lastToken.End != sourceLen {
		t.Errorf("last token end mismatch: expected %d, got %d",
			sourceLen, lastToken.End)
	}
}
