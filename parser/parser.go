package parser

import (
	"fmt"

	ts "github.com/tree-sitter/go-tree-sitter"
	"github.com/vietmpl/tree-sitter-vie/bindings/go"

	"github.com/vietmpl/vie/ast"
)

type parser struct {
	src []byte
	c   *ts.TreeCursor
}

var language = ts.NewLanguage(tree_sitter_vie.Language())

func ParseFile(src []byte) (*ast.SourceFile, error) {
	tsParser := ts.NewParser()
	tsParser.SetLanguage(language)
	defer tsParser.Close()

	tree := tsParser.Parse(src, nil)
	defer tree.Close()

	cursor := tree.Walk()
	defer cursor.Close()

	p := parser{
		src: src,
		c:   cursor,
	}

	blocks, err := p.blocks()
	if err != nil {
		return nil, err
	}
	return &ast.SourceFile{
		Blocks: blocks,
	}, nil
}

func (p parser) blocks() ([]ast.Block, error) {
	if !p.c.GotoFirstChild() {
		return nil, nil
	}
	var b []ast.Block
	for {
		bl, err := p.block()
		if err != nil {
			return nil, err
		}
		b = append(b, bl)
		if !p.c.GotoNextSibling() {
			break
		}
	}
	return b, nil
}

func (p parser) block() (ast.Block, error) {
	n := p.c.Node()
	switch n.Kind() {
	case "text_block":
		return &ast.TextBlock{
			Value: p.nodeContent(),
		}, nil
	case "render_block":
		p.c.GotoFirstChild()
		defer p.c.GotoParent()

		// Skip '{{'
		p.c.GotoNextSibling()

		expr, err := p.expr()
		if err != nil {
			return nil, err
		}

		return &ast.RenderBlock{
			Expr: expr,
		}, nil

	// case "if_block":
	// 	formatIf(node)
	// case "switch_block":
	// 	formatSwitch(node)
	default:
		panic(fmt.Sprintf("parser: unexpected node kind %s", n.Kind()))
	}
}

func (p parser) expr() (ast.Expr, error) {
	n := p.c.Node()
	switch n.Kind() {
	case "string_literal":
		return &ast.BasicLit{
			Kind:  ast.KindString,
			Value: p.nodeContent(),
		}, nil
	case "boolean_literal":
		return &ast.BasicLit{
			Kind:  ast.KindBool,
			Value: p.nodeContent(),
		}, nil
	case "identifier":
		return &ast.Ident{Value: p.nodeContent()}, nil

	case "unary_expression":
		p.c.GotoFirstChild()
		defer p.c.GotoParent()

		unary := &ast.UnaryExpr{}

		unary.Op = p.nodeContent()

		p.c.GotoNextSibling()

		expr, err := p.expr()
		if err != nil {
			return nil, err
		}

		unary.Expr = expr
		return unary, nil

	case "binary_expression":
		p.c.GotoFirstChild()
		defer p.c.GotoParent()

		binary := &ast.BinaryExpr{}

		left, err := p.expr()
		if err != nil {
			return nil, err
		}
		binary.Left = left

		p.c.GotoNextSibling()
		binary.Op = p.nodeContent()

		p.c.GotoNextSibling()
		right, err := p.expr()
		if err != nil {
			return nil, err
		}
		binary.Right = right

		return binary, nil

	case "call_expression":
		p.c.GotoFirstChild()
		defer p.c.GotoParent()
		call := &ast.CallExpr{}

		call.Func = &ast.Ident{Value: p.nodeContent()}

		// Skip '('
		p.c.GotoNextSibling()

		// Step into 'arguments'
		p.c.GotoFirstChild()
		defer p.c.GotoParent()

		for p.c.GotoNextSibling() {
			if p.c.Node().IsNamed() {
				expr, err := p.expr()
				if err != nil {
					return nil, err
				}

				call.Args = append(call.Args, expr)
			}
		}
		return call, nil

	case "pipe_expression":
		p.c.GotoFirstChild()
		defer p.c.GotoParent()
		pipe := &ast.PipeExpr{}

		arg, err := p.expr()
		if err != nil {
			return nil, err
		}
		pipe.Arg = arg

		// Skip '|'
		p.c.GotoNextSibling()

		p.c.GotoNextSibling()
		pipe.Func = &ast.Ident{Value: p.nodeContent()}
		return pipe, nil

	case "parenthesized_expression":
		p.c.GotoFirstChild()
		defer p.c.GotoParent()

		// Skip '('
		p.c.GotoNextSibling()

		expr, err := p.expr()
		if err != nil {
			return nil, err
		}

		return &ast.ParenExpr{
			Expr: expr,
		}, nil

	default:
		panic(fmt.Sprintf("parser: unexpected node kind %s", n.Kind()))
	}
}

func (p parser) nodeContent() []byte {
	n := p.c.Node()
	return p.src[n.StartByte():n.EndByte()]
}
