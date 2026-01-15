// Package render implements a rendering engine for Vie templates.
package render

import (
	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/value"
)

// Template renders a parsed Vie template using the provided data.
func Template(template *ast.Template, data map[string]value.Value) ([]byte, error) {
	r := renderer{
		data: data,
	}
	if err := r.renderBlocks(template.Blocks); err != nil {
		return nil, err
	}
	return r.buffer.Bytes(), nil
}

// TODO(skewb1k): consider adding a function that accepts an io.Writer as the
// output destination.
