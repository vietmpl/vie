package render

import (
	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/value"
)

func Template(template *ast.Template, data map[string]value.Value) ([]byte, error) {
	r := renderer{
		data: data,
	}
	if err := r.renderBlocks(template.Blocks); err != nil {
		return nil, err
	}
	return r.buffer.Bytes(), nil
}
