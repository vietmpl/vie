package render

import (
	"io"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/value"
)

func Template(w io.Writer, template *ast.Template, context map[string]value.Value) error {
	r := renderer{
		context: context,
		w:       w,
	}
	return r.renderBlocks(template.Blocks)
}
