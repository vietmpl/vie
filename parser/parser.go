package parser

import (
	"fmt"

	ts "github.com/tree-sitter/go-tree-sitter"
	"github.com/vietmpl/tree-sitter-vie/bindings/go"

	"github.com/vietmpl/vie/ast"
)

type parser struct {
	*ts.TreeCursor

	src []byte
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
		TreeCursor: cursor,
		src:        src,
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
	p.GotoFirstChild()
	defer p.GotoParent()
	var stmts []ast.Stmt
	for {
		stmt, err := p.stmt()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
		if !p.GotoNextSibling() {
			break
		}
	}
	return stmts, nil
}

func (p parser) stmt() (ast.Stmt, error) {
	n := p.Node()
	switch n.Kind() {
	case "text":
		return &ast.Text{
			Value: p.nodeContent(),
		}, nil

	case "render_statement":
		p.GotoFirstChild()
		defer p.GotoParent()

		// Consume '{{'
		p.GotoNextSibling()

		expr, err := p.expr()
		if err != nil {
			return nil, err
		}
		return &ast.RenderStmt{
			Expr: expr,
		}, nil

	case "if_statement":
		p.GotoFirstChild()
		defer p.GotoParent()
		ifStmt := &ast.IfStmt{}

		// Consume '{%'
		p.GotoNextSibling()
		// Consume 'if'
		p.GotoNextSibling()

		cond, err := p.expr()
		if err != nil {
			return nil, err
		}
		ifStmt.Condition = cond
		p.GotoNextSibling()

		// Consume '%}'
		p.GotoNextSibling()

		if p.FieldName() == "consequence" {
			consequence, err := p.stmts()
			if err != nil {
				return nil, err
			}
			ifStmt.Consequence = consequence
			p.GotoNextSibling()
		}

		if p.FieldName() == "alternative" {
			alt, err := p.alt()
			if err != nil {
				return nil, err
			}
			ifStmt.Alternative = alt
		}
		return ifStmt, nil

	case "switch_statement":
		p.GotoFirstChild()
		defer p.GotoParent()
		switchStmt := &ast.SwitchStmt{}

		// Consume '{%'
		p.GotoNextSibling()
		// Consume 'switch'
		p.GotoNextSibling()

		val, err := p.expr()
		if err != nil {
			return nil, err
		}
		switchStmt.Value = val
		p.GotoNextSibling()

		// Consume '%}'
		p.GotoNextSibling()

		for p.FieldName() == "cases" {
			caseClause, err := p.caseClause()
			if err != nil {
				return nil, err
			}
			switchStmt.Cases = append(switchStmt.Cases, caseClause)

			if !p.GotoNextSibling() {
				break
			}
		}
		return switchStmt, nil

	default:
		panic(fmt.Sprintf("parser: unexpected stmt kind %s", n.Kind()))
	}
}

func (p parser) caseClause() (*ast.CaseClause, error) {
	p.GotoFirstChild()
	defer p.GotoParent()
	caseClause := &ast.CaseClause{}

	// Consume '{%'
	p.GotoNextSibling()
	// Consume 'case'
	p.GotoNextSibling()

	val, err := p.expr()
	if err != nil {
		return nil, err
	}
	caseClause.Value = val
	p.GotoNextSibling()

	// Consume '%}'
	p.GotoNextSibling()

	if p.FieldName() == "body" {
		body, err := p.stmts()
		if err != nil {
			return nil, err
		}
		caseClause.Body = body
	}
	return caseClause, nil
}

func (p parser) alt() (any, error) {
	n := p.Node()
	switch n.Kind() {
	case "else_if_clause":
		p.GotoFirstChild()
		defer p.GotoParent()
		elseIf := &ast.ElseIfClause{}

		// Consume '{%'
		p.GotoNextSibling()
		// Consume 'else'
		p.GotoNextSibling()
		// Consume 'if'
		p.GotoNextSibling()

		cond, err := p.expr()
		if err != nil {
			return nil, err
		}
		elseIf.Condition = cond
		p.GotoNextSibling()

		// Consume '%}'
		p.GotoNextSibling()

		if p.FieldName() == "consequence" {
			consequence, err := p.stmts()
			if err != nil {
				return nil, err
			}
			elseIf.Consequence = consequence
		}

		if p.GotoNextSibling() {
			alt, err := p.alt()
			if err != nil {
				return nil, err
			}
			elseIf.Alternative = alt
		}

		return elseIf, nil

	case "else_clause":
		p.GotoFirstChild()
		defer p.GotoParent()

		// Consume '{%'
		p.GotoNextSibling()
		// Consume 'else'
		p.GotoNextSibling()
		// Consume '%}'
		if !p.GotoNextSibling() {
			return &ast.ElseClause{}, nil
		}

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
	n := p.Node()
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
		p.GotoFirstChild()
		defer p.GotoParent()
		unary := &ast.UnaryExpr{}

		unary.Op = ast.ParseUnOpKind(string(p.nodeContent()))

		p.GotoNextSibling()
		expr, err := p.expr()
		if err != nil {
			return nil, err
		}
		unary.Expr = expr

		return unary, nil

	case "binary_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		binary := &ast.BinaryExpr{}

		left, err := p.expr()
		if err != nil {
			return nil, err
		}
		binary.Left = left

		p.GotoNextSibling()
		binary.Op = ast.ParseBinOpKind(string(p.nodeContent()))

		p.GotoNextSibling()
		right, err := p.expr()
		if err != nil {
			return nil, err
		}
		binary.Right = right

		return binary, nil

	case "call_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		call := &ast.CallExpr{}

		call.Func = &ast.Ident{Value: p.nodeContent()}

		p.GotoNextSibling()

		p.GotoFirstChild()
		defer p.GotoParent()

		for {
			if p.Node().IsNamed() {
				expr, err := p.expr()
				if err != nil {
					return nil, err
				}
				call.Args = append(call.Args, expr)
			}

			if !p.GotoNextSibling() {
				break
			}
		}
		return call, nil

	case "pipe_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		pipe := &ast.PipeExpr{}

		arg, err := p.expr()
		if err != nil {
			return nil, err
		}
		pipe.Arg = arg

		// Consume '|'
		p.GotoNextSibling()

		p.GotoNextSibling()
		pipe.Func = &ast.Ident{Value: p.nodeContent()}

		return pipe, nil

	case "parenthesized_expression":
		p.GotoFirstChild()
		defer p.GotoParent()

		// Consume '('
		p.GotoNextSibling()

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
	n := p.Node()
	return p.src[n.StartByte():n.EndByte()]
}
