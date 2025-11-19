package parser

import (
	"bytes"
	"fmt"

	ts "github.com/tree-sitter/go-tree-sitter"
	"github.com/vietmpl/tree-sitter-vie/bindings/go"

	"github.com/vietmpl/vie/ast"
)

type parser struct {
	*ts.TreeCursor

	src  []byte
	path string
}

var language = ts.NewLanguage(tree_sitter_vie.Language())

func ParseBytes(src []byte, path string) (*ast.File, error) {
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
		path:       path,
	}

	var f ast.File
	if !p.GotoFirstChild() {
		return &f, nil
	}
	defer p.GotoParent()

	for {
		stmt := p.parseStmt()
		if stmt != nil {
			f.Stmts = append(f.Stmts, stmt)
		}
		if !p.GotoNextSibling() {
			break
		}
	}
	return &f, nil
}

func (p parser) parseStmt() ast.Stmt {
	n := p.Node()
	if n.IsMissing() {
		panic(fmt.Sprintf("parser: unexpected MISSING stmt %s", n.Kind()))
	}
	if n.IsError() {
		return &ast.BadStmt{
			From: p.posFromTsPoint(n.StartPosition()),
			To:   p.posFromTsPoint(n.EndPosition()),
		}
	}
	// TODO(skewb1k): use KindId instead of string comparisons.
	switch n.Kind() {
	case "text":
		b := p.src[n.StartByte():n.EndByte()]

		// TODO(skewb1k): improve performace.
		b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
		b = bytes.ReplaceAll(b, []byte("\r"), []byte("\n"))

		// TODO(skewb1k): factor out to parser.Peek().
		if p.GotoNextSibling() {
			// Trim trail spaces and tabs up to and including the first newline
			// if the previous node is a tag. 'text' nodes cannot follow
			// another 'text', so the next node must be a tag.
			if p.Node().Kind() != "render" && p.Node().Kind() != "comment_tag" {
				b = bytes.TrimRight(b, " \t")
				if len(b) > 0 && b[len(b)-1] != '\n' {
					b = append(b, '\n')
				}
			}
			p.GotoPreviousSibling()
		}
		if len(b) == 0 {
			return nil
		}
		return &ast.Text{
			Value: string(b),
		}

	case "comment_tag":
		var comment ast.Comment
		p.GotoFirstChild()
		defer p.GotoParent()

		p.GotoNextSibling() // '{#'
		// handle `{##}`
		commentNode := p.Node()
		if commentNode.IsError() {
			return &ast.BadStmt{
				From: p.posFromTsPoint(n.StartPosition()),
				// TODO(skewb1k): include final '#}' in the range.
				To: p.posFromTsPoint(commentNode.EndPosition()),
			}
		}
		if commentNode.Kind() == "comment" {
			comment.Content = p.nodeContent(p.Node())
		}
		return &comment

	case "render":
		var renderStmt ast.RenderStmt
		p.GotoFirstChild()
		defer p.GotoParent()

		p.GotoNextSibling() // '{{'
		renderStmt.X = p.parseExpr()
		p.GotoNextSibling() // '<expr>'
		// handle unexpected expression after first one
		// i.e. {{ "" "" }}
		nn := p.Node()
		if nn.IsError() {
			renderStmt.X = &ast.BadExpr{
				From: p.posFromTsPoint(nn.StartPosition()),
				To:   p.posFromTsPoint(nn.EndPosition()),
			}
		}
		return &renderStmt

	case "if_tag":
		bad := false
		var ifStmt ast.IfStmt
		p.GotoFirstChild()

		p.GotoNextSibling() // '{%'
		p.GotoNextSibling() // 'if'
		ifStmt.Cond = p.parseExpr()
		p.GotoParent()

		for p.GotoNextSibling() {
			switch p.Node().Kind() {
			case "else_if_tag":
				var elseIf ast.ElseIfClause
				p.GotoFirstChild()

				p.GotoNextSibling() // '{%'
				p.GotoNextSibling() // 'else'
				p.GotoNextSibling() // 'if'
				elseIf.Cond = p.parseExpr()
				from := p.Node().StartPosition()
				p.GotoNextSibling() // '<expr>'
				// handle unexpected expression after first one
				// i.e. {% else if "" "" %}
				nn := p.Node()
				if nn.IsError() {
					elseIf.Cond = &ast.BadExpr{
						From: p.posFromTsPoint(from),
						To:   p.posFromTsPoint(nn.EndPosition()),
					}
				}
				p.GotoParent()

				for p.GotoNextSibling() {
					kind := p.Node().Kind()
					if kind == "else_if_tag" || kind == "else_tag" || kind == "end_tag" {
						p.GotoPreviousSibling()
						break
					}
					stmt := p.parseStmt()
					if stmt != nil {
						elseIf.Cons = append(elseIf.Cons, stmt)
					}
				}
				ifStmt.ElseIfs = append(ifStmt.ElseIfs, elseIf)

			case "else_tag":
				var elseClause ast.ElseClause
				p.GotoFirstChild()

				p.GotoNextSibling() // '{%'
				p.GotoNextSibling() // 'else'
				// handle unexpected expression after else
				// i.e. {% else "" %}
				if p.Node().IsError() {
					bad = true
				}
				p.GotoParent()

				for p.GotoNextSibling() {
					if p.Node().Kind() == "end_tag" {
						p.GotoPreviousSibling()
						break
					}
					stmt := p.parseStmt()
					if stmt != nil {
						elseClause.Cons = append(elseClause.Cons, stmt)
					}
				}
				ifStmt.Else = &elseClause

			case "end_tag":
				if bad {
					return &ast.BadStmt{
						From: p.posFromTsPoint(n.StartPosition()),
						To:   p.posFromTsPoint(p.Node().EndPosition()),
					}
				}
				return &ifStmt

			default:
				stmt := p.parseStmt()
				if stmt != nil {
					ifStmt.Cons = append(ifStmt.Cons, stmt)
				}
			}
		}
		// TODO(skewb1k): restore TSCursor to the last valid node rather than
		// advancing to EOF when an end_tag is missing.
		return &ast.BadStmt{
			From: p.posFromTsPoint(n.StartPosition()),
			To:   p.posFromTsPoint(p.Node().EndPosition()),
		}

	case "switch_tag":
		bad := false
		var switchStmt ast.SwitchStmt
		p.GotoFirstChild()

		p.GotoNextSibling() // '{%'
		p.GotoNextSibling() // 'switch'
		switchStmt.Value = p.parseExpr()
		p.GotoParent()

		for p.GotoNextSibling() {
			switch p.Node().Kind() {
			case "case_tag":
				var caseClause ast.CaseClause
				p.GotoFirstChild()

				p.GotoNextSibling() // '{%'
				p.GotoNextSibling() // 'case'
				// handle unexpected expression after case
				// i.e. {% case "" "" %}
				if p.Node().IsError() {
					bad = true
					p.GotoNextSibling() // '<bad-expr>'
				}
				caseClause.List = p.parseExprList()
				p.GotoNextSibling() // '<expr-list>'
				p.GotoParent()

				for p.GotoNextSibling() {
					k := p.Node().Kind()
					if k == "case_tag" || k == "end_tag" {
						p.GotoPreviousSibling()
						break
					}
					stmt := p.parseStmt()
					if stmt != nil {
						caseClause.Body = append(caseClause.Body, stmt)
					}
				}
				switchStmt.Cases = append(switchStmt.Cases, caseClause)

			case "end_tag":
				if bad {
					return &ast.BadStmt{
						From: p.posFromTsPoint(n.StartPosition()),
						To:   p.posFromTsPoint(p.Node().EndPosition()),
					}
				}
				return &switchStmt

			case "text":
				// TODO(skewb1k): allow only whitespaces.

			default:
				bad = true
			}
		}
		// TODO(skewb1k): restore TSCursor to the last valid node rather than
		// advancing to EOF when an end_tag is missing.
		return &ast.BadStmt{
			From: p.posFromTsPoint(n.StartPosition()),
			To:   p.posFromTsPoint(n.EndPosition()),
		}

	case "end_tag", "else_if_tag", "else_tag", "case_tag":
		return &ast.BadStmt{
			From: p.posFromTsPoint(n.StartPosition()),
			To:   p.posFromTsPoint(n.EndPosition()),
		}

	default:
		panic(fmt.Sprintf("parser: unexpected stmt kind %q while parsing %s", n.Kind(), p.src))
	}
}

func (p parser) parseExpr() ast.Expr {
	n := p.Node()
	if n.IsError() || n.IsMissing() {
		return &ast.BadExpr{
			From: p.posFromTsPoint(n.StartPosition()),
			To:   p.posFromTsPoint(n.EndPosition()),
		}
	}
	switch n.Kind() {
	case "string_literal":
		return &ast.BasicLit{
			ValuePos: p.posFromTsPoint(n.StartPosition()),
			Kind:     ast.KindString,
			Value:    p.nodeContent(n),
		}

	case "boolean_literal":
		return &ast.BasicLit{
			ValuePos: p.posFromTsPoint(n.StartPosition()),
			Kind:     ast.KindBool,
			Value:    p.nodeContent(n),
		}

	case "identifier":
		return &ast.Ident{
			NamePos: p.posFromTsPoint(n.StartPosition()),
			Name:    p.nodeContent(n),
		}

	case "unary_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var unary ast.UnaryExpr

		nn := p.Node()
		unary.OpPos = p.posFromTsPoint(nn.StartPosition())
		unary.Op = ast.ParseUnOpKind(string(p.nodeContent(nn)))

		p.GotoNextSibling()
		unary.X = p.parseExpr()

		return &unary

	case "binary_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var binary ast.BinaryExpr

		binary.X = p.parseExpr()

		p.GotoNextSibling()
		binary.Op = ast.ParseBinOpKind(string(p.nodeContent(p.Node())))

		p.GotoNextSibling()
		binary.Y = p.parseExpr()

		return &binary

	case "call_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var call ast.CallExpr

		nn := p.Node()
		call.Func = ast.Ident{
			NamePos: p.posFromTsPoint(nn.StartPosition()),
			Name:    p.nodeContent(nn),
		}
		p.GotoNextSibling()

		call.Args = p.parseExprList()
		return &call

	case "pipe_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var pipe ast.PipeExpr

		pipe.Arg = p.parseExpr()
		p.GotoNextSibling() // <expr>

		p.GotoNextSibling() // '|'

		nn := p.Node()
		if nn.IsError() || nn.IsMissing() {
			return &ast.BadExpr{
				From: p.posFromTsPoint(n.StartPosition()),
				To:   p.posFromTsPoint(nn.EndPosition()),
			}
		}

		pipe.Func = ast.Ident{Name: p.nodeContent(p.Node())}

		return &pipe

	case "parenthesized_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var paren ast.ParenExpr

		paren.Lparen = p.posFromTsPoint(p.Node().StartPosition())
		p.GotoNextSibling()

		paren.X = p.parseExpr()
		return &paren

	// case "render","end_tag", "else_if_tag", "else_tag", "case_tag":

	// TODO(skewb1k): ideally, we should panic when encountering an
	// unrecognized TSKind, instead of silently producing a placeholder.
	default:
		return &ast.BadExpr{
			From: p.posFromTsPoint(n.StartPosition()),
			To:   p.posFromTsPoint(n.EndPosition()),
		}
		// panic(fmt.Sprintf("parser: unexpected expr kind %q while parsing %s", n.Kind(), p.src))
	}
}

func (p parser) parseExprList() []ast.Expr {
	p.GotoFirstChild()
	defer p.GotoParent()

	var list []ast.Expr
	for {
		if p.Node().IsNamed() {
			list = append(list, p.parseExpr())
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

func (p parser) posFromTsPoint(point ts.Point) ast.Pos {
	return ast.Pos{
		Path:      p.path,
		Line:      point.Row,
		Character: point.Column,
	}
}
