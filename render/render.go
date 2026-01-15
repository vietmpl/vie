package render

import (
	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/value"
)

func Template(template *ast.Template, context map[string]value.Value) ([]byte, error) {
	r := renderer{
		context: context,
	}
	if err := r.renderBlocks(template.Blocks); err != nil {
		return nil, err
	}
	return r.buffer.Bytes(), nil
}
