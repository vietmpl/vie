package format

import (
	"bytes"
	"io"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/parse"
)

// Template formats input file and writes the result to dst.
//
// The function may return early (before the entire result is written)
// and return a formatting error, for instance due to an incorrect AST.
func Template(dst io.Writer, template *ast.Template) error {
	f := formatter{
		w: dst,
	}
	f.blocks(template.Blocks)
	return nil
}

// Source is a convenience function that formats src and returns the result or
// an error in case of a syntax error. src is expecterd to be a a syntactically
// correct Vie Template file.
//
// It buffers the entire formatted result internally; callers that want to
// stream output directly could use [Template] instead.
func Source(src []byte) ([]byte, error) {
	parsed, err := parse.Source(src)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := Template(&buf, parsed); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
