package parser

import (
	"bytes"
	"fmt"
	"strings"

	ts "github.com/tree-sitter/go-tree-sitter"
	"github.com/vietmpl/tree-sitter-vie/bindings/go"

	"github.com/vietmpl/vie/ast"
)

var vieLanguage = ts.NewLanguage(tree_sitter_vie.Language())

type parser struct {
	*ts.TreeCursor

	src []byte
}

func ParseBytes(src []byte) (*ast.Template, error) {
	tsParser := ts.NewParser()
	_ = tsParser.SetLanguage(vieLanguage)
	defer tsParser.Close()

	tree := tsParser.Parse(src, nil)
	defer tree.Close()

	cursor := tree.Walk()
	defer cursor.Close()

	p := parser{
		TreeCursor: cursor,
		src:        src,
	}

	var template ast.Template
	if !p.GotoFirstChild() {
		// the file is empty.
		return &template, nil
	}
	defer p.GotoParent()

	for {
		block, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		if block != nil {
			template.Blocks = append(template.Blocks, block)
		}
		if !p.GotoNextSibling() {
			break
		}
	}
	return &template, nil
}

func (p *parser) parseBlock() (ast.Block, error) {
	n := p.Node()
	if n.IsError() {
		return nil, fmt.Errorf("invalid block")
	}
	// TODO(skewb1k): use KindId instead of string comparisons.
	switch n.Kind() {
	case "text":
		b := p.src[n.StartByte():n.EndByte()]

		// TODO(skewb1k): improve performace.
		b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
		b = bytes.ReplaceAll(b, []byte("\r"), []byte("\n"))

		if len(b) == 0 {
			return nil, nil
		}
		return &ast.TextBlock{
			Content: string(b),
		}, nil

	case "comment_tag":
		var comment ast.CommentBlock
		p.GotoFirstChild()
		defer p.GotoParent()

		p.GotoNextSibling() // '{#'
		// handle `{##}`
		commentNode := p.Node()
		if commentNode.IsError() {
			return nil, fmt.Errorf("comments cannot contain line breaks")
		}
		if commentNode.Kind() == "comment" {
			comment.Content = p.nodeContent(p.Node())
		}
		return &comment, nil

	case "render":
		var displayBlock ast.DisplayBlock
		p.GotoFirstChild()
		defer p.GotoParent()

		p.GotoNextSibling() // '{{'
		value, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		displayBlock.Value = value
		p.GotoNextSibling() // '<expr>'
		// handle `{{ "" "" }}`
		nn := p.Node()
		if nn.IsError() {
			return nil, fmt.Errorf("unexpected %s in display statement", p.nodeContent(nn))
		}
		return &displayBlock, nil

	case "if_tag":
		var ifBlock ast.IfBlock
		p.GotoFirstChild()

		p.GotoNextSibling() // '{%'
		p.GotoNextSibling() // 'if'
		condition, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		ifBlock.Branches = append(ifBlock.Branches, ast.IfBranch{
			Condition: condition,
		})
		p.GotoParent()

		for p.GotoNextSibling() {
			switch p.Node().Kind() {
			case "elseif_tag":
				var elseIf ast.IfBranch
				p.GotoFirstChild()

				p.GotoNextSibling() // '{%'
				p.GotoNextSibling() // 'elseif'
				condition, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				elseIf.Condition = condition
				p.GotoParent()

				for p.GotoNextSibling() {
					kind := p.Node().Kind()
					if kind == "elseif_tag" || kind == "else_tag" || kind == "end_tag" {
						p.GotoPreviousSibling()
						break
					}
					block, err := p.parseBlock()
					if err != nil {
						return nil, err
					}
					if block != nil {
						elseIf.Consequence = append(elseIf.Consequence, block)
					}
				}
				ifBlock.Branches = append(ifBlock.Branches, elseIf)

			case "else_tag":
				p.GotoFirstChild()

				p.GotoNextSibling() // '{%'
				p.GotoNextSibling() // 'else'
				// handle `{% else "" %}`
				nn := p.Node()
				if nn.IsError() {
					content := p.nodeContent(nn)
					return nil, fmt.Errorf("unexpected %q after else", content)
				}
				p.GotoParent()

				ifBlock.ElseConsequence = &[]ast.Block{}
				for p.GotoNextSibling() {
					if p.Node().Kind() == "end_tag" {
						p.GotoPreviousSibling()
						break
					}
					block, err := p.parseBlock()
					if err != nil {
						return nil, err
					}
					if block != nil {
						*ifBlock.ElseConsequence = append(*ifBlock.ElseConsequence, block)
					}
				}

			case "end_tag":
				return &ifBlock, nil

			default:
				block, err := p.parseBlock()
				if err != nil {
					return nil, err
				}
				if block != nil {
					ifBlock.Branches[0].Consequence = append(ifBlock.Branches[0].Consequence, block)
				}
			}
		}

		return nil, fmt.Errorf("expected {%% end %%}, found EOF")

	case "end_tag", "elseif_tag", "else_tag":
		return nil, fmt.Errorf("unexpected %s", strings.TrimSpace(p.nodeContent(n)))

	default:
		panic(fmt.Sprintf("parser: unexpected block kind %q while parsing %s", n.Kind(), p.src))
	}
}

func (p *parser) parseExpr() (ast.Expr, error) {
	n := p.Node()
	if n.IsError() || n.IsMissing() {
		return nil, fmt.Errorf("expected expression, found %s", p.nodeContent(n))
	}
	switch n.Kind() {
	case "string_literal":
		return &ast.BasicLiteral{
			Start_: posFromTsPoint(n.StartPosition()),
			Kind:   ast.KindString,
			Value:  p.nodeContent(n),
		}, nil

	case "boolean_literal":
		return &ast.BasicLiteral{
			Start_: posFromTsPoint(n.StartPosition()),
			Kind:   ast.KindBool,
			Value:  p.nodeContent(n),
		}, nil

	case "identifier":
		return &ast.Identifier{
			Start_: posFromTsPoint(n.StartPosition()),
			Value:  p.nodeContent(n),
		}, nil

	case "unary_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var unary ast.UnaryExpr

		nn := p.Node()
		unary.OperatorLocation = posFromTsPoint(nn.StartPosition())
		unary.Operator = ast.ParseUnaryOperator(p.nodeContent(nn))

		p.GotoNextSibling()
		operand, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		unary.Operand = operand
		return &unary, nil

	case "binary_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var binary ast.BinaryExpr

		lOperand, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		binary.LOperand = lOperand

		p.GotoNextSibling()
		binary.Operator = ast.ParseBinaryOperator(p.nodeContent(p.Node()))

		p.GotoNextSibling()
		rOperand, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		binary.ROperand = rOperand

		return &binary, nil

	case "call_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var call ast.CallExpr

		nn := p.Node()
		call.Function = ast.Identifier{
			Start_: posFromTsPoint(nn.StartPosition()),
			Value:  p.nodeContent(nn),
		}
		p.GotoNextSibling()
		arguments, err := p.parseExprList()
		if err != nil {
			return nil, err
		}
		call.Arguments = arguments

		return &call, nil

	case "pipe_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var pipe ast.PipeExpr

		argument, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		pipe.Argument = argument
		p.GotoNextSibling() // <expr>

		p.GotoNextSibling() // '|'

		nn := p.Node()
		if nn.IsError() || nn.IsMissing() {
			return nil, fmt.Errorf("expected expression, found %s", p.nodeContent(n))
		}
		pipe.Function = ast.Identifier{Value: p.nodeContent(nn)}

		return &pipe, nil

	case "parenthesized_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var paren ast.ParenExpr

		paren.LparenLocation = posFromTsPoint(p.Node().StartPosition())
		p.GotoNextSibling()

		value, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		paren.Value = value
		return &paren, nil

	default:
		return nil, fmt.Errorf("expected expression, found %s", p.nodeContent(n))
	}
}

func (p *parser) parseExprList() ([]ast.Expr, error) {
	p.GotoFirstChild()
	defer p.GotoParent()

	var list []ast.Expr
	for {
		if p.Node().IsNamed() {
			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			list = append(list, expr)
		}
		if !p.GotoNextSibling() {
			break
		}
	}
	return list, nil
}

func (p *parser) nodeContent(n *ts.Node) string {
	return string(p.src[n.StartByte():n.EndByte()])
}

func posFromTsPoint(point ts.Point) ast.Location {
	return ast.Location{
		Line:   point.Row,
		Column: point.Column,
	}
}
