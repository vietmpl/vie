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

	stmts, err := p.stmts()
	if err != nil {
		return nil, err
	}
	return &ast.SourceFile{
		Stmts: stmts,
	}, nil
}

func (p parser) stmts() ([]ast.Stmt, error) {
	if !p.c.GotoFirstChild() {
		return nil, nil
	}
	defer p.c.GotoParent()
	var stmts []ast.Stmt
	for {
		stmt, err := p.stmt()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
		if !p.c.GotoNextSibling() {
			break
		}
	}
	return stmts, nil
}

func (p parser) stmt() (ast.Stmt, error) {
	n := p.c.Node()
	switch n.Kind() {
	case "text":
		return &ast.Text{
			Value: p.nodeContent(),
		}, nil

	case "render_statement":
		p.c.GotoFirstChild()
		defer p.c.GotoParent()

		// Consume '{{'
		p.c.GotoNextSibling()

		expr, err := p.expr()
		if err != nil {
			return nil, err
		}
		return &ast.RenderStmt{
			Expr: expr,
		}, nil

	case "if_statement":
		p.c.GotoFirstChild()
		defer p.c.GotoParent()
		ifstmt := &ast.IfStmt{}

		// Consume '{%'
		p.c.GotoNextSibling()
		// Consume 'if'
		p.c.GotoNextSibling()

		cond, err := p.expr()
		if err != nil {
			return nil, err
		}
		ifstmt.Condition = cond
		p.c.GotoNextSibling()

		// Consume '%}'
		p.c.GotoNextSibling()

		if p.c.FieldName() == "consequence" {
			consequence, err := p.stmts()
			if err != nil {
				return nil, err
			}
			ifstmt.Consequence = consequence
			p.c.GotoNextSibling()
		}

		if p.c.FieldName() == "alternative" {
			alt, err := p.alt()
			if err != nil {
				return nil, err
			}
			ifstmt.Alternative = alt
		}

		return ifstmt, nil

	// case "switch_block":
	default:
		panic(fmt.Sprintf("parser: unexpected block kind %s", n.Kind()))
	}
}

func (p parser) alt() (any, error) {
	n := p.c.Node()
	switch n.Kind() {
	case "else_if_clause":
		p.c.GotoFirstChild()
		defer p.c.GotoParent()
		elseIf := &ast.ElseIfClause{}

		// Consume '{%'
		p.c.GotoNextSibling()
		// Consume 'else'
		p.c.GotoNextSibling()
		// Consume 'if'
		p.c.GotoNextSibling()

		cond, err := p.expr()
		if err != nil {
			return nil, err
		}
		elseIf.Condition = cond
		p.c.GotoNextSibling()

		// Consume '%}'
		p.c.GotoNextSibling()

		consequence, err := p.stmts()
		if err != nil {
			return nil, err
		}
		elseIf.Consequence = consequence

		if p.c.GotoNextSibling() {
			alt, err := p.alt()
			if err != nil {
				return nil, err
			}
			elseIf.Alternative = alt
		}

		return elseIf, nil

	case "else_clause":
		p.c.GotoFirstChild()
		defer p.c.GotoParent()

		// Consume '{%'
		p.c.GotoNextSibling()
		// Consume 'end'
		p.c.GotoNextSibling()
		// Consume '%}'
		p.c.GotoNextSibling()

		consequence, err := p.stmts()
		if err != nil {
			return nil, err
		}
		return &ast.ElseClause{
			Consequence: consequence,
		}, nil

	default:
		panic(fmt.Sprintf("parser: unexpected alt kind %s", n.Kind()))
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

		p.c.GotoNextSibling()

		p.c.GotoFirstChild()
		defer p.c.GotoParent()

		for {
			if p.c.Node().IsNamed() {
				expr, err := p.expr()
				if err != nil {
					return nil, err
				}
				call.Args = append(call.Args, expr)
			}

			if !p.c.GotoNextSibling() {
				break
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

		// Consume '|'
		p.c.GotoNextSibling()

		p.c.GotoNextSibling()
		pipe.Func = &ast.Ident{Value: p.nodeContent()}

		return pipe, nil

	case "parenthesized_expression":
		p.c.GotoFirstChild()
		defer p.c.GotoParent()

		// Consume '('
		p.c.GotoNextSibling()

		expr, err := p.expr()
		if err != nil {
			return nil, err
		}
		return &ast.ParenExpr{
			Expr: expr,
		}, nil

	default:
		panic(fmt.Sprintf("parser: unexpected expr kind %s", n.Kind()))
	}
}

func (p parser) nodeContent() []byte {
	n := p.c.Node()
	return p.src[n.StartByte():n.EndByte()]
}
