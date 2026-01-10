package parse

import (
	"fmt"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/lexer"
	"github.com/vietmpl/vie/token"
)

type parser struct {
	source []byte
	index  int
	tokens []lexer.Token
}

func (p *parser) peek() lexer.Token {
	return p.tokens[p.index]
}

func (p *parser) next() lexer.Token {
	tok := p.tokens[p.index]
	p.index++
	return tok
}

func (p *parser) tokenContent(tok lexer.Token) string {
	return string(p.source[tok.Start:tok.End])
}

// TODO: consider returning token.
func (p *parser) expectToken(kind token.Kind) error {
	currentToken := p.next()
	if currentToken.Kind != kind {
		return fmt.Errorf("expected %s, got %s",
			kind,
			currentToken.Kind)
	}
	return nil
}

func Source(source []byte) (*ast.Template, error) {
	p := parser{
		source: source,
	}
	var l lexer.Lexer
	l.Init(source)
	for {
		tok := l.Next()
		p.tokens = append(p.tokens, tok)
		if tok.Kind == token.EOF {
			break
		}
	}

	var blocks []ast.Block
	var block ast.Block
	var err error

loop:
	for {
		currentToken := p.next()
		switch currentToken.Kind {
		case token.EOF:
			break loop
		case token.TEXT:
			block = p.parseTextBlock(currentToken)
		case token.L_DOUBLE_BRACE:
			block, err = p.parseDisplayBlock()
		case token.L_BRACE_PERCENT:
			block, err = p.parseStmt()
		case token.L_BRACE_POUND:
			block, err = p.parseCommentBlock()
		default:
			return nil, fmt.Errorf("unexpected token: %s", currentToken.Kind)
		}

		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return &ast.Template{
		Blocks: blocks,
	}, nil
}

func (p *parser) parseTextBlock(tok lexer.Token) *ast.TextBlock {
	return &ast.TextBlock{
		Content: p.tokenContent(tok),
	}
}

func (p *parser) parseCommentBlock() (*ast.CommentBlock, error) {
	currentToken := p.next()
	switch currentToken.Kind {
	case token.COMMENT:
		content := p.tokenContent(currentToken)
		if err := p.expectToken(token.R_BRACE_POUND); err != nil {
			return nil, err
		}
		return &ast.CommentBlock{
			Content: content,
		}, nil
	case token.R_BRACE_POUND:
		return &ast.CommentBlock{}, nil
	default:
		return nil, fmt.Errorf("unexpected token: %s", currentToken.Kind)
	}
}

func (p *parser) parseDisplayBlock() (*ast.DisplayBlock, error) {
	value, err := p.parseExpr(0)
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(token.R_DOUBLE_BRACE); err != nil {
		return nil, err
	}

	return &ast.DisplayBlock{
		Value: value,
	}, nil
}

func (p *parser) parseStmt() (ast.Block, error) {
	currentToken := p.next()
	switch currentToken.Kind {
	case token.KEYWORD_IF:
		return p.parseIfBlock()
	default:
		return nil, fmt.Errorf("unexpected token: %s", currentToken.Kind)
	}
}

func (p *parser) parseIfBlock() (*ast.IfBlock, error) {
	var (
		branches []ast.IfBranch
		elseBody []ast.Block
		inElse   bool
	)

	cond, err := p.parseExpr(0)
	if err != nil {
		return nil, err
	}
	if err := p.expectToken(token.R_BRACE_PERCENT); err != nil {
		return nil, err
	}

	branches = append(branches, ast.IfBranch{
		Condition:   cond,
		Consequence: nil,
	})

	currentBranch := &branches[0]

	for {
		currentToken := p.next()
		switch currentToken.Kind {
		case token.TEXT:
			block := p.parseTextBlock(currentToken)
			if inElse {
				elseBody = append(elseBody, block)
			} else {
				currentBranch.Consequence = append(currentBranch.Consequence, block)
			}

		case token.L_DOUBLE_BRACE:
			block, err := p.parseDisplayBlock()
			if err != nil {
				return nil, err
			}
			if inElse {
				elseBody = append(elseBody, block)
			} else {
				currentBranch.Consequence = append(currentBranch.Consequence, block)
			}

		case token.L_BRACE_PERCENT:
			currentToken := p.next()
			switch currentToken.Kind {
			case token.KEYWORD_ELSEIF:
				if inElse {
					return nil, fmt.Errorf("unexpected elseif tag after else")
				}

				cond, err := p.parseExpr(0)
				if err != nil {
					return nil, err
				}
				if err := p.expectToken(token.R_BRACE_PERCENT); err != nil {
					return nil, err
				}

				branches = append(branches, ast.IfBranch{
					Condition:   cond,
					Consequence: nil,
				})
				currentBranch = &branches[len(branches)-1]

			case token.KEYWORD_ELSE:
				if err := p.expectToken(token.R_BRACE_PERCENT); err != nil {
					return nil, err
				}
				inElse = true

			case token.KEYWORD_END:
				if err := p.expectToken(token.R_BRACE_PERCENT); err != nil {
					return nil, err
				}

				var elsePtr *[]ast.Block
				if inElse {
					elsePtr = &elseBody
				}

				return &ast.IfBlock{
					Branches:    branches,
					Alternative: elsePtr,
				}, nil

			default:
				return nil, fmt.Errorf("unexpected token in if block: %s", currentToken.Kind)
			}

		default:
			return nil, fmt.Errorf("unexpected token: %s", currentToken.Kind)
		}
	}
}

var precMap = map[token.Kind]struct {
	prec  int
	assoc bool
}{
	token.KEYWORD_OR:  {prec: 1, assoc: false},
	token.KEYWORD_AND: {prec: 2, assoc: false},
	token.EQUAL_EQUAL: {prec: 3, assoc: false},
	token.PIPE:        {prec: 7, assoc: false},
}

func (p *parser) parseExpr(minPrec int) (ast.Expr, error) {
	node, err := p.parsePrimaryExpr()
	if err != nil {
		return nil, err
	}

	for {
		currentToken := p.peek()
		// TODO(skewb1k): fixme.
		if currentToken.Kind == token.R_DOUBLE_BRACE || currentToken.Kind == token.R_BRACE_PERCENT {
			break
		}
		op, ok := precMap[currentToken.Kind]
		if !ok {
			return nil, fmt.Errorf("unexpected token: %s", currentToken.Kind)
		}
		if op.prec < minPrec {
			break
		}
		p.index++

		right, err := p.parseExpr(op.prec)
		if err != nil {
			return nil, err
		}

		node = &ast.BinaryExpr{
			LOperand: node,
			Operator: 0,
			ROperand: right,
		}
	}

	return node, nil
}

func (p *parser) parsePrimaryExpr() (ast.Expr, error) {
	currentToken := p.next()
	switch currentToken.Kind {
	case token.IDENTIFIER:
		return &ast.Identifier{
			// Start_: ast.Pos{},
			Value: p.tokenContent(currentToken),
		}, nil

	case token.BANG:
		operand, err := p.parsePrimaryExpr()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{
			// OperatorLocation: ast.Location{},
			Operator: token.BANG,
			Operand:  operand,
		}, nil

	case token.L_PAREN:
		value, err := p.parseExpr(0)
		if err != nil {
			return nil, err
		}
		if err := p.expectToken(token.R_PAREN); err != nil {
			return nil, err
		}
		return &ast.ParenExpr{
			// LparenLocation: ast.Location{},
			Value: value,
		}, nil

	default:
		return nil, fmt.Errorf("expected expression, got %s", currentToken.Kind)
	}
}
