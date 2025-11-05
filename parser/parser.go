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
	_ = tsParser.SetLanguage(language)
	defer tsParser.Close()

	tree := tsParser.Parse(src, nil)
	defer tree.Close()

	cursor := tree.Walk()
	defer cursor.Close()

	p := parser{
		TreeCursor: cursor,
		src:        src,
	}

	p.GotoFirstChild()
	defer p.GotoParent()

	var sf ast.SourceFile
	for {
		stmt := p.stmt()
		sf.Stmts = append(sf.Stmts, stmt)
		if !p.GotoNextSibling() {
			break
		}
	}
	return &sf
}

func (p parser) stmt() ast.Stmt {
	n := p.Node()
	switch n.Kind() {
	case "text":
		return &ast.Text{
			Value: p.nodeContent(n),
		}

	case "render":
		p.GotoFirstChild()
		defer p.GotoParent()

		p.GotoNextSibling() // '{{'

		return &ast.RenderStmt{
			X: p.expr(),
		}

	case "if_tag":
		var ifStmt ast.IfStmt
		p.GotoFirstChild()

		p.GotoNextSibling() // '{%'
		p.GotoNextSibling() // 'if'
		ifStmt.Cond = p.expr()
		p.GotoParent()

		for p.GotoNextSibling() {
			switch p.Node().Kind() {
			case "else_if_tag":
				var elseIf ast.ElseIfClause
				p.GotoFirstChild()

				p.GotoNextSibling() // '{%'
				p.GotoNextSibling() // 'else'
				p.GotoNextSibling() // 'if'
				elseIf.Cond = p.expr()
				p.GotoParent()

				for p.GotoNextSibling() {
					kind := p.Node().Kind()
					if kind == "else_if_tag" || kind == "else_tag" || kind == "end_tag" {
						p.GotoPreviousSibling()
						break
					}
					elseIf.Cons = append(elseIf.Cons, p.stmt())
				}
				ifStmt.ElseIfs = append(ifStmt.ElseIfs, elseIf)

			case "else_tag":
				var elseClause ast.ElseClause
				p.GotoFirstChild()

				p.GotoNextSibling() // '{%'
				p.GotoNextSibling() // 'else'
				p.GotoParent()

				for p.GotoNextSibling() {
					if p.Node().Kind() == "end_tag" {
						p.GotoPreviousSibling()
						break
					}
					elseClause.Cons = append(elseClause.Cons, p.stmt())
				}
				ifStmt.Else = &elseClause

			case "end_tag":
				return &ifStmt

			default:
				ifStmt.Cons = append(ifStmt.Cons, p.stmt())
			}
		}
		panic("parser: unexpected EOF when parsing if_tag")

	case "switch_tag":
		var switchStmt ast.SwitchStmt
		p.GotoFirstChild()

		p.GotoNextSibling() // '{%'
		p.GotoNextSibling() // 'switch'
		switchStmt.Value = p.expr()
		p.GotoParent()

		for p.GotoNextSibling() {
			switch p.Node().Kind() {
			case "case_tag":
				var caseClause ast.CaseClause
				p.GotoFirstChild()

				p.GotoNextSibling() // '{%'
				p.GotoNextSibling() // 'case'
				caseClause.List = p.exprList()
				p.GotoParent()

				for p.GotoNextSibling() {
					k := p.Node().Kind()
					if k == "case_tag" || k == "end_tag" {
						p.GotoPreviousSibling()
						break
					}
					caseClause.Body = append(caseClause.Body, p.stmt())
				}
				switchStmt.Cases = append(switchStmt.Cases, caseClause)

			case "end_tag":
				return &switchStmt

			case "text":
				// TODO: allow only whitespaces.

			default:
				panic(fmt.Sprintf("parser: unexpected tag when parsing switch kind %s", p.Node().Kind()))
			}
		}
		panic("parser: unexpected EOF when parsing switch_tag")

	default:
		panic(fmt.Sprintf("parser: unexpected stmt kind %s", n.Kind()))
	}
}

func (p parser) expr() ast.Expr {
	n := p.Node()
	switch n.Kind() {
	case "string_literal":
		return &ast.BasicLit{
			ValuePos: posFromTsPoint(n.StartPosition()),
			Kind:     ast.KindString,
			Value:    p.nodeContent(n),
		}

	case "boolean_literal":
		return &ast.BasicLit{
			ValuePos: posFromTsPoint(n.StartPosition()),
			Kind:     ast.KindBool,
			Value:    p.nodeContent(n),
		}

	case "identifier":
		return &ast.Ident{
			NamePos: posFromTsPoint(n.StartPosition()),
			Name:    p.nodeContent(n),
		}

	case "unary_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var unary ast.UnaryExpr

		n := p.Node()
		unary.OpPos = posFromTsPoint(n.StartPosition())
		unary.Op = ast.ParseUnOpKind(string(p.nodeContent(n)))

		p.GotoNextSibling()
		unary.X = p.expr()

		return &unary

	case "binary_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var binary ast.BinaryExpr

		binary.X = p.expr()

		p.GotoNextSibling()
		binary.Op = ast.ParseBinOpKind(string(p.nodeContent(p.Node())))

		p.GotoNextSibling()
		binary.Y = p.expr()

		return &binary

	case "call_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var call ast.CallExpr

		n := p.Node()
		call.Func = ast.Ident{
			NamePos: posFromTsPoint(n.StartPosition()),
			Name:    p.nodeContent(n),
		}
		p.GotoNextSibling()

		call.Args = p.exprList()
		return &call

	case "pipe_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var pipe ast.PipeExpr

		pipe.Arg = p.expr()

		p.GotoNextSibling() // '|'

		p.GotoNextSibling()
		pipe.Func = ast.Ident{Name: p.nodeContent(p.Node())}

		return &pipe

	case "parenthesized_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var paren ast.ParenExpr

		paren.Lparen = posFromTsPoint(p.Node().StartPosition())
		p.GotoNextSibling()

		paren.X = p.expr()
		return &paren

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

func (p parser) nodeContent(n *ts.Node) string {
	return string(p.src[n.StartByte():n.EndByte()])
}

func posFromTsPoint(p ts.Point) ast.Pos {
	return ast.Pos{
		Line:      p.Row,
		Character: p.Column,
	}
}
