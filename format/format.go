package format

import (
	"fmt"
	"io"

	ts "github.com/tree-sitter/go-tree-sitter"
	"github.com/vietmpl/tree-sitter-vie/bindings/go"
)

var language = ts.NewLanguage(tree_sitter_vie.Language())

func Source(w io.Writer, src []byte) error {
	tsParser := ts.NewParser()
	tsParser.SetLanguage(language)
	defer tsParser.Close()

	tree := tsParser.Parse(src, nil)
	defer tree.Close()

	c := tree.Walk()
	defer c.Close()

	if !c.GotoFirstChild() {
		return nil
	}

	for {
		n := c.Node()
		switch n.Kind() {
		case "text":
			if _, err := w.Write(nodeContent(src, n)); err != nil {
				return err
			}
		case "render_block":
			c.GotoFirstChild()
			// Skip '{{'
			c.GotoNextSibling()
			if _, err := w.Write([]byte("{{ ")); err != nil {
				return err
			}
			if err := formatExpr(w, src, c); err != nil {
				return err
			}
			if _, err := w.Write([]byte(" }}")); err != nil {
				return err
			}
			c.GotoParent()
		// case "if_block":
		// 	formatIf(node)
		// case "switch_block":
		// 	formatSwitch(node)
		default:
			panic(fmt.Sprintf("unexpected node kind %s", n.Kind()))
		}

		if !c.GotoNextSibling() {
			return nil
		}
	}
}

func formatExpr(w io.Writer, src []byte, c *ts.TreeCursor) error {
	n := c.Node()
	switch n.Kind() {
	case "identifier", "string_literal", "boolean_literal":
		if _, err := w.Write(nodeContent(src, n)); err != nil {
			return err
		}

	case "unary_expression":
		c.GotoFirstChild()
		defer c.GotoParent()
		op := nodeContent(src, c.Node())
		if _, err := w.Write(op); err != nil {
			return err
		}
		// do not insert whitespace after '!'
		if string(op) != "!" {
			if _, err := w.Write([]byte{' '}); err != nil {
				return err
			}
		}

		c.GotoNextSibling()
		if err := formatExpr(w, src, c); err != nil {
			return err
		}

	case "binary_expression":
		c.GotoFirstChild()
		defer c.GotoParent()
		if err := formatExpr(w, src, c); err != nil {
			return err
		}
		if _, err := w.Write([]byte{' '}); err != nil {
			return err
		}

		c.GotoNextSibling()
		if _, err := w.Write(nodeContent(src, c.Node())); err != nil {
			return err
		}
		if _, err := w.Write([]byte{' '}); err != nil {
			return err
		}

		c.GotoNextSibling()
		if err := formatExpr(w, src, c); err != nil {
			return err
		}

	case "call_expression":
		c.GotoFirstChild()
		defer c.GotoParent()
		if _, err := w.Write(nodeContent(src, c.Node())); err != nil {
			return err
		}

		// Step into 'arguments'
		c.GotoNextSibling()
		c.GotoFirstChild()
		defer c.GotoParent()

		// Skip '('
		if _, err := w.Write([]byte{'('}); err != nil {
			return err
		}

		// TODO(skewb1k): optimize and remove duplication
		for c.GotoNextSibling() && c.Node().Kind() != ")" {
			if err := formatExpr(w, src, c); err != nil {
				return err
			}
			// peek ahead
			if c.GotoNextSibling() && c.Node().Kind() != ")" {
				if _, err := w.Write([]byte(", ")); err != nil {
					return err
				}
				continue
			}
			break
		}

		if _, err := w.Write([]byte{')'}); err != nil {
			return err
		}

	case "pipe_expression":
		c.GotoFirstChild()
		defer c.GotoParent()
		if err := formatExpr(w, src, c); err != nil {
			return err
		}

		// Skip '|'
		c.GotoNextSibling()
		if _, err := w.Write([]byte(" | ")); err != nil {
			return err
		}

		c.GotoNextSibling()
		if err := formatExpr(w, src, c); err != nil {
			return err
		}

	default:
		panic(fmt.Sprintf("unexpected node kind %s", n.Kind()))
	}
	return nil
}

func nodeContent(src []byte, n *ts.Node) []byte {
	return src[n.StartByte():n.EndByte()]
}
