package template

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/vietmpl/vie/render"
	"github.com/vietmpl/vie/value"
)

func (t Template) Render(context map[string]value.Value) (map[string][]byte, error) {
	files := make(map[string][]byte)
	var buf bytes.Buffer

	onFile := func(f *File, parent string) error {
		render.MustRenderFile(&buf, f.NameAST, context)
		name := buf.String()
		buf.Reset()
		path := filepath.Join(parent, name)

		if f.ContentAST != nil {
			path = strings.TrimSuffix(path, ".vie")
			render.MustRenderFile(&buf, f.ContentAST, context)
			if _, ok := files[path]; ok {
				return fmt.Errorf("%s conflicts", path)
			}
			files[path] = buf.Bytes()
			buf.Reset()
		} else {
			files[path] = f.Content
		}
		return nil
	}

	onDir := func(d *Dir, parent string) error {
		render.MustRenderFile(&buf, d.NameAST, context)
		name := buf.String()
		buf.Reset()
		d.Name = name
		return nil
	}

	if err := t.Walk(onDir, onFile); err != nil {
		return nil, err
	}
	return files, nil
}
