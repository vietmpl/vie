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
			Value: p.nodeContent(p.Node()),
		}, nil

	case "render_statement":
		p.GotoFirstChild()
		defer p.GotoParent()

		// Consume '{{'
		p.GotoNextSibling()

		x, err := p.expr()
		if err != nil {
			return nil, err
		}
		return &ast.RenderStmt{
			X: x,
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
		ifStmt.Cond = cond
		p.GotoNextSibling()

		// Consume '%}'
		p.GotoNextSibling()

		if p.FieldName() == "consequence" {
			cons, err := p.stmts()
			if err != nil {
				return nil, err
			}
			ifStmt.Cons = cons
			p.GotoNextSibling()
		}

		if p.FieldName() == "alternative" {
			alt, err := p.alt()
			if err != nil {
				return nil, err
			}
			ifStmt.Alt = alt
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

	list, err := p.exprList()
	if err != nil {
		return nil, err
	}
	caseClause.List = list
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
		elseIf.Cond = cond
		p.GotoNextSibling()

		// Consume '%}'
		p.GotoNextSibling()

		if p.FieldName() == "consequence" {
			cons, err := p.stmts()
			if err != nil {
				return nil, err
			}
			elseIf.Cons = cons
		}

		if p.GotoNextSibling() {
			alt, err := p.alt()
			if err != nil {
				return nil, err
			}
			elseIf.Alt = alt
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

		cons, err := p.stmts()
		if err != nil {
			return nil, err
		}
		return &ast.ElseClause{
			Cons: cons,
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
			ValuePos: fromTsPoint(n.StartPosition()),
			Kind:     ast.KindString,
			Value:    p.nodeContent(p.Node()),
		}, nil

	case "boolean_literal":
		return &ast.BasicLit{
			ValuePos: fromTsPoint(n.StartPosition()),
			Kind:     ast.KindBool,
			Value:    p.nodeContent(p.Node()),
		}, nil

	case "identifier":
		return &ast.Ident{
			NamePos: fromTsPoint(n.StartPosition()),
			Name:    p.nodeContent(p.Node()),
		}, nil

	case "unary_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		unary := &ast.UnaryExpr{}

		n := p.Node()
		unary.OpPos = fromTsPoint(n.StartPosition())
		unary.Op = ast.ParseUnOpKind(string(p.nodeContent(n)))

		p.GotoNextSibling()
		x, err := p.expr()
		if err != nil {
			return nil, err
		}
		unary.X = x

		return unary, nil

	case "binary_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		binary := &ast.BinaryExpr{}

		x, err := p.expr()
		if err != nil {
			return nil, err
		}
		binary.X = x

		p.GotoNextSibling()
		binary.Op = ast.ParseBinOpKind(string(p.nodeContent(p.Node())))

		p.GotoNextSibling()
		y, err := p.expr()
		if err != nil {
			return nil, err
		}
		binary.Y = y

		return binary, nil

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

		args, err := p.exprList()
		if err != nil {
			return nil, err
		}
		call.Args = args
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
		pipe.Func = &ast.Ident{Name: p.nodeContent(p.Node())}

		return pipe, nil

	case "parenthesized_expression":
		p.GotoFirstChild()
		defer p.GotoParent()
		paren := &ast.ParenExpr{
			Lparen: fromTsPoint(p.Node().StartPosition()),
		}

		// Consume '('
		p.GotoNextSibling()

		x, err := p.expr()
		if err != nil {
			return nil, err
		}
		paren.X = x
		return paren, nil

	default:
		panic(fmt.Sprintf("parser: unexpected expr kind %s", n.Kind()))
	}
}

func (p parser) exprList() ([]ast.Expr, error) {
	p.GotoFirstChild()
	defer p.GotoParent()

	var list []ast.Expr
	for {
		if p.Node().IsNamed() {
			expr, err := p.expr()
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

func (p parser) nodeContent(n *ts.Node) []byte {
	return p.src[n.StartByte():n.EndByte()]
}

func fromTsPoint(p ts.Point) ast.Pos {
	return ast.Pos{
		Line:      p.Row,
		Character: p.Column,
	}
}
