// Package format provides canonical formatting for Vie template source code.
package format

import (
	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/parse"
)

// Template formats parsed Vie template and returns the result.
func Template(template *ast.Template) []byte {
	var p printer
	p.printBlocks(template.Blocks)
	return p.buffer.Bytes()
}

// Source formats raw Vie template source and returns the result or a syntax error.
//
// For already parsed templates, use [Template].
func Source(src []byte) ([]byte, error) {
	template, err := parse.Source(src)
	if err != nil {
		return nil, err
	}

	return Template(template), nil
}
