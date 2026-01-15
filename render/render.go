package render

import (
	"bytes"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/value"
)

func Template(template *ast.Template, context map[string]value.Value) ([]byte, error) {
	var buf bytes.Buffer
	r := renderer{
		context: context,
		w:       &buf,
	}
	if err := r.renderBlocks(template.Blocks); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
