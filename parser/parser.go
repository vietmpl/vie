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

func ParseFile(src []byte) *ast.SourceFile {
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

	stmts := p.stmts()
	return &ast.SourceFile{
		Stmts: stmts,
	}
}

func (p parser) stmts() []ast.Stmt {
	p.GotoFirstChild()
	defer p.GotoParent()
	var stmts []ast.Stmt
	for {
		stmt := p.stmt()
		stmts = append(stmts, stmt)
		if !p.GotoNextSibling() {
			break
		}
	}
	return stmts
}

func (p parser) stmt() ast.Stmt {
	n := p.Node()
	switch n.Kind() {
	case "text":
		return &ast.Text{
			Value: p.nodeContent(p.Node()),
		}

	case "render_statement":
		p.GotoFirstChild()
		defer p.GotoParent()

		// Consume '{{'
		p.GotoNextSibling()

		return &ast.RenderStmt{
			X: p.expr(),
		}

	case "if_statement":
		p.GotoFirstChild()
		defer p.GotoParent()
		ifStmt := &ast.IfStmt{}

		// Consume '{%'
		p.GotoNextSibling()
		// Consume 'if'
		p.GotoNextSibling()

		ifStmt.Cond = p.expr()
		p.GotoNextSibling()

		// Consume '%}'
		p.GotoNextSibling()

		if p.FieldName() == "consequence" {
			ifStmt.Cons = p.stmts()
			p.GotoNextSibling()
		}
		if p.FieldName() == "alternative" {
			ifStmt.Alt = p.alt()
		}
		return ifStmt

	case "switch_statement":
		p.GotoFirstChild()
		defer p.GotoParent()
		switchStmt := &ast.SwitchStmt{}

		// Consume '{%'
		p.GotoNextSibling()
		// Consume 'switch'
		p.GotoNextSibling()

		switchStmt.Value = p.expr()
		p.GotoNextSibling()

		// Consume '%}'
		p.GotoNextSibling()

		for p.FieldName() == "cases" {
			switchStmt.Cases = append(switchStmt.Cases, p.caseClause())
			if !p.GotoNextSibling() {
				break
			}
		}
		return switchStmt

	default:
		panic(fmt.Sprintf("parser: unexpected stmt kind %s", n.Kind()))
	}
}

func (p parser) caseClause() *ast.CaseClause {
	p.GotoFirstChild()
	defer p.GotoParent()
	caseClause := &ast.CaseClause{}

	// Consume '{%'
	p.GotoNextSibling()
	// Consume 'case'
	p.GotoNextSibling()

	caseClause.List = p.exprList()
	p.GotoNextSibling()

	// Consume '%}'
	p.GotoNextSibling()

	if p.FieldName() == "body" {
		caseClause.Body = p.stmts()
	}
	return caseClause
}

func (p parser) alt() any {
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

		elseIf.Cond = p.expr()
		p.GotoNextSibling()

		// Consume '%}'
		p.GotoNextSibling()

		if p.FieldName() == "consequence" {
			elseIf.Cons = p.stmts()
		}

		if p.GotoNextSibling() {
			elseIf.Alt = p.alt()
		}

		return elseIf

	case "else_clause":
		p.GotoFirstChild()
		defer p.GotoParent()

		// Consume '{%'
		p.GotoNextSibling()
		// Consume 'else'
		p.GotoNextSibling()
		// Consume '%}'
		if p.GotoNextSibling() {
			cons := p.stmts()
			return &ast.ElseClause{
				Cons: cons,
			}
		}
		return &ast.ElseClause{}

	default:
		panic(fmt.Sprintf("parser: unexpected alt kind %s", n.Kind()))
	}
}

func (p parser) expr() ast.Expr {
	n := p.Node()
	switch n.Kind() {
	case "string_literal":
		return &ast.BasicLit{
			ValuePos: fromTsPoint(n.StartPosition()),
			Kind:     ast.KindString,
			Value:    p.nodeContent(p.Node()),
		}

	case "boolean_literal":
		return &ast.BasicLit{
			ValuePos: fromTsPoint(n.StartPosition()),
			Kind:     ast.KindBool,
			Value:    p.nodeContent(p.Node()),
		}

	case "identifier":
		return &ast.Ident{
			NamePos: fromTsPoint(n.StartPosition()),
			Name:    p.nodeContent(p.Node()),
		}

	case "unary_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		unary := &ast.UnaryExpr{}

		n := p.Node()
		unary.OpPos = fromTsPoint(n.StartPosition())
		unary.Op = ast.ParseUnOpKind(string(p.nodeContent(n)))

		p.GotoNextSibling()
		unary.X = p.expr()

		return unary

	case "binary_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		binary := &ast.BinaryExpr{}

		binary.X = p.expr()

		p.GotoNextSibling()
		binary.Op = ast.ParseBinOpKind(string(p.nodeContent(p.Node())))

		p.GotoNextSibling()
		binary.Y = p.expr()

		return binary

	case "call_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		call := &ast.CallExpr{}

		n := p.Node()
		call.Fn = &ast.Ident{
			NamePos: fromTsPoint(n.StartPosition()),
			Name:    p.nodeContent(n),
		}
		p.GotoNextSibling()

		call.Args = p.exprList()
		return call

	case "pipe_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		pipe := &ast.PipeExpr{}

		pipe.Arg = p.expr()

		// Consume '|'
		p.GotoNextSibling()

		p.GotoNextSibling()
		pipe.Func = &ast.Ident{Name: p.nodeContent(p.Node())}

		return pipe

	case "parenthesized_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		paren := &ast.ParenExpr{
			Lparen: fromTsPoint(p.Node().StartPosition()),
		}

		// Consume '('
		p.GotoNextSibling()

		paren.X = p.expr()
		return paren

	default:
		panic(fmt.Sprintf("parser: unexpected expr kind %s", n.Kind()))
	}
}

func (p parser) exprList() []ast.Expr {
	p.GotoFirstChild()
	defer p.GotoParent()

	var list []ast.Expr
	for {
		if p.Node().IsNamed() {
			list = append(list, p.expr())
		}
		if !p.GotoNextSibling() {
			break
		}
	}
	return list
}

func (p parser) nodeContent(n *ts.Node) []byte {
	return p.src[n.StartByte():n.EndByte()]
}

func fromTsPoint(p ts.Point) ast.Pos {
	return ast.Pos{
		Line:      p.Row,
		Character: p.Column,
	}
}
