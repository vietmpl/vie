package format

import (
	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/parse"
)

func Template(template *ast.Template) []byte {
	var f formatter
	f.blocks(template.Blocks)
	return f.buffer.Bytes()
}

// Source is a convenience function that formats src and returns the result or
// an error in case of a syntax error.
func Source(src []byte) ([]byte, error) {
	template, err := parse.Source(src)
	if err != nil {
		return nil, err
	}

	return Template(template), nil
}
