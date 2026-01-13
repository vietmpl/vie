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
		if err := render.Template(&buf, f.NameTemplate, context); err != nil {
			return err
		}
		name := buf.String()
		buf.Reset()
		path := filepath.Join(parent, name)

		if f.ContentTemplate != nil {
			path = strings.TrimSuffix(path, ".vie")
			if err := render.Template(&buf, f.ContentTemplate, context); err != nil {
				return err
			}
			bytes := append([]byte(nil), buf.Bytes()...) // copy
			buf.Reset()
			if _, ok := files[path]; ok {
				return fmt.Errorf("%s conflicts", path)
			}
			files[path] = bytes
		} else {
			files[path] = f.Content
		}
		return nil
	}

	onDir := func(d *Dir, parent string) error {
		if err := render.Template(&buf, d.NameTemplate, context); err != nil {
			return err
		}
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
