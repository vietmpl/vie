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

	src    []byte
	errors ErrorList
}

func ParseBytes(src []byte) (*ast.File, error) {
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

	var f ast.File
	if !p.GotoFirstChild() {
		// the file is empty.
		return &f, nil
	}
	defer p.GotoParent()

	for {
		block := p.parseBlock()
		if block != nil {
			f.Blocks = append(f.Blocks, block)
		}
		if !p.GotoNextSibling() {
			break
		}
	}
	var err error
	if p.errors.Len() > 0 {
		err = p.errors
	}
	return &f, err
}

func (p *parser) parseBlock() ast.Block {
	n := p.Node()
	if n.IsError() {
		from := posFromTsPoint(n.StartPosition())
		p.errors.Add(from, "invalid statement")
		return &ast.BadBlock{
			From: from,
			To:   posFromTsPoint(n.EndPosition()),
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
		return &ast.TextBlock{
			Value: string(b),
		}

	case "comment_tag":
		var comment ast.CommentBlock
		p.GotoFirstChild()
		defer p.GotoParent()

		p.GotoNextSibling() // '{#'
		// handle `{##}`
		commentNode := p.Node()
		if commentNode.IsError() {
			return p.addBadBlockAndError(
				n.StartPosition(),
				commentNode.EndPosition(),
				"comments cannot contain line breaks",
			)
		}
		if commentNode.Kind() == "comment" {
			comment.Content = p.nodeContent(p.Node())
		}
		return &comment

	case "render":
		var renderBlock ast.RenderBlock
		p.GotoFirstChild()
		defer p.GotoParent()

		p.GotoNextSibling() // '{{'
		from := p.Node().StartPosition()
		renderBlock.X = p.parseExpr()
		p.GotoNextSibling() // '<expr>'
		// handle `{{ "" "" }}`
		nn := p.Node()
		if nn.IsError() {
			msg := fmt.Sprintf("unexpected %s in render statement", p.nodeContent(nn))
			p.errors.Add(posFromTsPoint(nn.StartPosition()), msg)
			renderBlock.X = &ast.BadExpr{
				From: posFromTsPoint(from),
				// TODO(skewb1k): include all nodes left.
				To: posFromTsPoint(nn.EndPosition()),
			}
		}
		return &renderBlock

	case "if_tag":
		bad := false
		var ifBlock ast.IfBlock
		p.GotoFirstChild()

		p.GotoNextSibling() // '{%'
		p.GotoNextSibling() // 'if'
		ifBlock.Cond = p.parseExpr()
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
				p.GotoParent()

				for p.GotoNextSibling() {
					kind := p.Node().Kind()
					if kind == "else_if_tag" || kind == "else_tag" || kind == "end_tag" {
						p.GotoPreviousSibling()
						break
					}
					block := p.parseBlock()
					if block != nil {
						elseIf.Cons = append(elseIf.Cons, block)
					}
				}
				ifBlock.ElseIfs = append(ifBlock.ElseIfs, elseIf)

			case "else_tag":
				var elseClause ast.ElseClause
				p.GotoFirstChild()

				p.GotoNextSibling() // '{%'
				p.GotoNextSibling() // 'else'
				// handle `{% else "" %}`
				nn := p.Node()
				if nn.IsError() {
					content := p.nodeContent(nn)
					var msg string
					if content == "if" {
						msg = "missing condition in else if statement"
					} else {
						msg = fmt.Sprintf("unexpected %q after else", content)
					}
					p.errors.Add(posFromTsPoint(n.StartPosition()), msg)
					bad = true
				}
				p.GotoParent()

				for p.GotoNextSibling() {
					if p.Node().Kind() == "end_tag" {
						p.GotoPreviousSibling()
						break
					}
					block := p.parseBlock()
					if block != nil {
						elseClause.Cons = append(elseClause.Cons, block)
					}
				}
				ifBlock.Else = &elseClause

			case "end_tag":
				if bad {
					return &ast.BadBlock{
						From: posFromTsPoint(n.StartPosition()),
						To:   posFromTsPoint(p.Node().EndPosition()),
					}
				}
				return &ifBlock

			default:
				block := p.parseBlock()
				if block != nil {
					ifBlock.Cons = append(ifBlock.Cons, block)
				}
			}
		}
		// TODO(skewb1k): restore TSCursor to the last valid node rather than
		// advancing to EOF when an end_tag is missing.
		return p.addBadBlockAndError(
			n.StartPosition(),
			p.Node().EndPosition(),
			"expected {% end %}, found EOF",
		)

	case "switch_tag":
		bad := false
		var switchBlock ast.SwitchBlock
		p.GotoFirstChild()

		p.GotoNextSibling() // '{%'
		p.GotoNextSibling() // 'switch'
		switchBlock.Value = p.parseExpr()
		p.GotoParent()

		for p.GotoNextSibling() {
			switch p.Node().Kind() {
			case "case_tag":
				var caseClause ast.CaseClause
				p.GotoFirstChild()

				p.GotoNextSibling() // '{%'
				p.GotoNextSibling() // 'case'
				nn := p.Node()
				if nn.IsError() {
					from := posFromTsPoint(nn.StartPosition())
					p.errors.Add(from, fmt.Sprintf("expected expression, found %s", p.nodeContent(nn)))
					caseClause.List = []ast.Expr{
						&ast.BadExpr{
							From: from,
							To:   posFromTsPoint(nn.EndPosition()),
						},
					}
				} else {
					caseClause.List = p.parseExprList()
				}
				p.GotoParent()

				for p.GotoNextSibling() {
					k := p.Node().Kind()
					if k == "case_tag" || k == "end_tag" {
						p.GotoPreviousSibling()
						break
					}
					block := p.parseBlock()
					if block != nil {
						caseClause.Body = append(caseClause.Body, block)
					}
				}
				switchBlock.Cases = append(switchBlock.Cases, caseClause)

			case "end_tag":
				if bad {
					return &ast.BadBlock{
						From: posFromTsPoint(n.StartPosition()),
						To:   posFromTsPoint(p.Node().EndPosition()),
					}
				}
				return &switchBlock

			case "text":
				// TODO(skewb1k): allow only whitespaces.

			default:
				bad = true
			}
		}
		// TODO(skewb1k): restore TSCursor to the last valid node rather than
		// advancing to EOF when an end_tag is missing.
		return p.addBadBlockAndError(
			n.StartPosition(),
			p.Node().EndPosition(),
			"expected {% end %}, found EOF",
		)

	case "end_tag", "else_if_tag", "else_tag", "case_tag":
		return p.addBadBlockAndError(
			n.StartPosition(),
			n.EndPosition(),
			// tag nodes may contain trailing \n.
			fmt.Sprintf("unexpected %s", strings.TrimSpace(p.nodeContent(n))),
		)

	default:
		panic(fmt.Sprintf("parser: unexpected block kind %q while parsing %s", n.Kind(), p.src))
	}
}

func (p *parser) parseExpr() ast.Expr {
	n := p.Node()
	if n.IsError() || n.IsMissing() {
		from := posFromTsPoint(n.StartPosition())
		p.errors.Add(from, fmt.Sprintf("expected expression, found %s", p.nodeContent(n)))
		return &ast.BadExpr{
			From: from,
			To:   posFromTsPoint(n.EndPosition()),
		}
	}
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

		nn := p.Node()
		unary.OpPos = posFromTsPoint(nn.StartPosition())
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
			NamePos: posFromTsPoint(nn.StartPosition()),
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
			from := posFromTsPoint(n.StartPosition())
			p.errors.Add(from, fmt.Sprintf("expected expression, found %s", p.nodeContent(n)))
			return &ast.BadExpr{
				From: from,
				To:   posFromTsPoint(nn.EndPosition()),
			}
		}

		pipe.Func = ast.Ident{Name: p.nodeContent(p.Node())}

		return &pipe

	case "parenthesized_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		var paren ast.ParenExpr

		paren.Lparen = posFromTsPoint(p.Node().StartPosition())
		p.GotoNextSibling()

		paren.X = p.parseExpr()
		return &paren

	// case "render","end_tag", "else_if_tag", "else_tag", "case_tag":

	// TODO(skewb1k): ideally, we should panic when encountering an
	// unrecognized TSKind, instead of silently producing a placeholder.
	default:
		from := posFromTsPoint(n.StartPosition())
		p.errors.Add(from, fmt.Sprintf("expected expression, found %s", p.nodeContent(n)))
		return &ast.BadExpr{
			From: from,
			To:   posFromTsPoint(n.EndPosition()),
		}
		// panic(fmt.Sprintf("parser: unexpected expr kind %q while parsing %s", n.Kind(), p.src))
	}
}

func (p *parser) parseExprList() []ast.Expr {
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

func (p *parser) nodeContent(n *ts.Node) string {
	return string(p.src[n.StartByte():n.EndByte()])
}

func posFromTsPoint(point ts.Point) ast.Pos {
	return ast.Pos{
		Line:      point.Row,
		Character: point.Column,
	}
}
func (p *parser) addBadBlockAndError(from ts.Point, to ts.Point, msg string) *ast.BadBlock {
	f := posFromTsPoint(from)
	p.errors.Add(f, msg)
	return &ast.BadBlock{
		From: f,
		To:   posFromTsPoint(to),
	}

}
