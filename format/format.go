package format

import (
	"bytes"
	"io"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/parser"
)

// File formats input file and writes the result to dst.
//
// The function may return early (before the entire result is written)
// and return a formatting error, for instance due to an incorrect AST.
func File(dst io.Writer, file *ast.File) error {
	f := formatter{
		w: dst,
	}
	f.stmts(file.Stmts)
	return nil
}

// Source is a convenience function that formats src and returns the result or
// an error in case of a syntax error. src is expecterd to be a a syntactically
// correct Vie Template file.
//
// It buffers the entire formatted result internally; callers that want to
// stream output directly could use [File] instead.
func Source(src []byte) ([]byte, error) {
	parsed, err := parser.ParseBytes(src)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := File(&buf, parsed); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
